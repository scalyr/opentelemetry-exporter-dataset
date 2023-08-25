/*
 * Copyright 2023 SentinelOne, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/scalyr/dataset-go/pkg/server_host_config"

	"golang.org/x/exp/slices"

	"github.com/scalyr/dataset-go/pkg/version"

	"github.com/cenkalti/backoff/v4"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/buffer"
	"github.com/scalyr/dataset-go/pkg/config"

	_ "net/http/pprof"

	"github.com/cskr/pubsub"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	HttpErrorCannotConnect   = 600
	HttpErrorHasErrorMessage = 499
)

// isOkStatus returns true if status code is 200, false otherwise.
func isOkStatus(status uint32) bool {
	return status == http.StatusOK
}

// isRetryableStatus returns true if status code is 401, 403, 429, or any 5xx, false otherwise.
// Note from DataSet API request handling perspective 401 and 403 may be considered as retryable, see DSET-2201
func isRetryableStatus(status uint32) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

// DataSetClient represent a DataSet REST API client
type DataSetClient struct {
	Id          uuid.UUID
	Config      *config.DataSetConfig
	Client      *http.Client
	SessionInfo *add_events.SessionInfo
	// map of known Buffer //TODO introduce cleanup
	buffers          map[string]*buffer.Buffer
	buffersAllMutex  sync.Mutex
	buffersEnqueued  atomic.Uint64
	buffersProcessed atomic.Uint64
	buffersDropped   atomic.Uint64
	buffersBroken    atomic.Uint64
	// Pub/Sub topics of Buffers based on its session
	BufferPerSessionTopic *pubsub.PubSub
	LastHttpStatus        atomic.Uint32
	lastError             error
	lastErrorTs           time.Time
	lastErrorMu           sync.RWMutex
	retryAfter            time.Time
	retryAfterMu          sync.RWMutex
	// indicates that client has been shut down and no further processing is possible
	finished         atomic.Bool
	Logger           *zap.Logger
	eventsEnqueued   atomic.Uint64
	eventsProcessed  atomic.Uint64
	eventsDropped    atomic.Uint64
	eventsBroken     atomic.Uint64
	bytesAPISent     atomic.Uint64
	bytesAPIAccepted atomic.Uint64
	addEventsMutex   sync.Mutex
	// Pub/Sub topics of EventBundles based on its key
	eventBundlePerKeyTopic *pubsub.PubSub
	// map of known Subscription channels of eventBundlePerKeyTopic
	eventBundleSubscriptionChannels map[string]chan interface{}
	// timestamp of first event processed by client
	firstReceivedAt atomic.Int64
	// timestamp of last event so far processed by client
	lastAcceptedAt atomic.Int64
	// Stores sanitized complete URL to the addEvents API endpoint, e.g.
	// https://app.scalyr.com/api/addEvents
	addEventsEndpointUrl string
	userAgent            string
	serverHost           string
}

func NewClient(cfg *config.DataSetConfig, client *http.Client, logger *zap.Logger, userAgentSuffix *string) (*DataSetClient, error) {
	logger.Info(
		"Using config: ",
		zap.String("config", cfg.String()),
		zap.String("version", version.Version),
		zap.String("releaseDate", version.ReleasedDate),
	)

	validationErr := cfg.Validate()
	if validationErr != nil {
		return nil, validationErr
	}

	// update group by, so that logs from the same host
	// belong to the same session
	addServerHostIntoGroupBy(cfg)

	serverHost, err := getServerHost(cfg.ServerHostSettings)
	if err != nil {
		return nil, fmt.Errorf("it was not get server host: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("it was not possible to generate UUID: %w", err)
	}

	addEventsEndpointUrl := cfg.Endpoint
	if strings.HasSuffix(addEventsEndpointUrl, "/") {
		addEventsEndpointUrl += "api/addEvents"
	} else {
		addEventsEndpointUrl += "/api/addEvents"
	}

	userAgent := fmt.Sprintf(
		"%s;%s;%s;%s;%s;%s;%d",
		"dataset-go",
		version.Version,
		version.ReleasedDate,
		id,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.NumCPU(),
	)
	if userAgentSuffix != nil && *userAgentSuffix != "" {
		userAgent = userAgent + ";" + *userAgentSuffix
	}

	dataClient := &DataSetClient{
		Id:                              id,
		Config:                          cfg,
		Client:                          client,
		SessionInfo:                     &add_events.SessionInfo{},
		buffers:                         make(map[string]*buffer.Buffer),
		buffersEnqueued:                 atomic.Uint64{},
		buffersProcessed:                atomic.Uint64{},
		buffersDropped:                  atomic.Uint64{},
		buffersBroken:                   atomic.Uint64{},
		buffersAllMutex:                 sync.Mutex{},
		BufferPerSessionTopic:           pubsub.New(0),
		LastHttpStatus:                  atomic.Uint32{},
		retryAfter:                      time.Now(),
		retryAfterMu:                    sync.RWMutex{},
		lastErrorMu:                     sync.RWMutex{},
		Logger:                          logger,
		finished:                        atomic.Bool{},
		eventsEnqueued:                  atomic.Uint64{},
		eventsProcessed:                 atomic.Uint64{},
		eventsDropped:                   atomic.Uint64{},
		eventsBroken:                    atomic.Uint64{},
		bytesAPIAccepted:                atomic.Uint64{},
		bytesAPISent:                    atomic.Uint64{},
		addEventsMutex:                  sync.Mutex{},
		eventBundlePerKeyTopic:          pubsub.New(0),
		eventBundleSubscriptionChannels: make(map[string]chan interface{}),
		firstReceivedAt:                 atomic.Int64{},
		lastAcceptedAt:                  atomic.Int64{},
		addEventsEndpointUrl:            addEventsEndpointUrl,
		userAgent:                       userAgent,
		serverHost:                      serverHost,
	}

	// run buffer sweeper if requested
	if cfg.BufferSettings.MaxLifetime > 0 {
		dataClient.Logger.Info("Buffer.MaxLifetime is positive => send buffers regularly",
			zap.Duration("Buffer.MaxLifetime", cfg.BufferSettings.MaxLifetime),
		)
		go dataClient.bufferSweeper(cfg.BufferSettings.MaxLifetime)
	} else {
		dataClient.Logger.Warn(
			"Buffer.MaxLifetime is not positive => do NOT send buffers regularly",
			zap.Duration("Buffer.MaxLifetime", cfg.BufferSettings.MaxLifetime),
		)
	}

	// run statistics sweeper
	go dataClient.statisticsSweeper()

	dataClient.Logger.Info("DataSetClient was created",
		zap.String("id", dataClient.Id.String()),
	)

	//// Server for pprof
	//go func() {
	//	fmt.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	return dataClient, nil
}

func getServerHost(settings server_host_config.DataSetServerHostSettings) (string, error) {
	err := settings.Validate()
	if err != nil {
		return "", err
	}
	if len(settings.ServerHost) > 0 {
		return settings.ServerHost, nil
	}
	return os.Hostname()
}

// addServerHostIntoGroupBy adds attributes that indicate from which machine
// the logs are into the groupBy attribute, so that they are part of the same session
func addServerHostIntoGroupBy(cfg *config.DataSetConfig) {
	groupBy := cfg.BufferSettings.GroupBy
	if !slices.Contains(groupBy, add_events.AttrOrigServerHost) {
		groupBy = append(groupBy, add_events.AttrOrigServerHost)
	}
	cfg.BufferSettings.GroupBy = groupBy
}

func (client *DataSetClient) getBuffer(key string) *buffer.Buffer {
	session := fmt.Sprintf("%s-%s", client.Id, key)
	client.buffersAllMutex.Lock()
	defer client.buffersAllMutex.Unlock()
	return client.buffers[session]
}

func (client *DataSetClient) initBuffer(buff *buffer.Buffer, info *add_events.SessionInfo) {
	client.Logger.Debug("Creating new buf",
		zap.String("uuid", buff.Id.String()),
		zap.String("session", buff.Session),
	)

	// Initialise
	err := buff.Initialise(info)
	// only reason why Initialise may fail is because SessionInfo cannot be
	// serialised into JSON. We do not care, so we can set it to nil and continue
	if err != nil {
		client.Logger.Error("Cannot initialize buffer: %w", zap.Error(err))
		err = buff.SetSessionInfo(nil)
		if err != nil {
			panic(fmt.Sprintf("Setting SessionInfo cannot cause an error: %s", err))
		}
	}
}

func (client *DataSetClient) newBuffersSubscriberRoutine(session string) {
	ch := client.BufferPerSessionTopic.Sub(session)
	go (func(session string, ch chan interface{}) {
		client.listenAndSendBufferForSession(session, ch)
	})(session, ch)
}

func (client *DataSetClient) listenAndSendBufferForSession(session string, ch chan interface{}) {
	client.Logger.Info("Listening to submit buffer",
		zap.String("session", session),
	)

	for processedMsgCnt := 0; ; processedMsgCnt++ {
		msg, channelReceiveSuccess := <-ch
		if !channelReceiveSuccess {
			client.Logger.Error(
				"Cannot receive Buffer from channel",
				zap.String("session", session),
				zap.Any("msg", msg),
			)
			client.buffersBroken.Add(1)
			client.lastAcceptedAt.Store(time.Now().UnixNano())
			continue
		}
		client.Logger.Debug("Received Buffer from channel",
			zap.String("session", session),
			zap.Int("processedMsgCnt", processedMsgCnt),
			zap.Uint64("buffersEnqueued", client.buffersEnqueued.Load()),
			zap.Uint64("buffersProcessed", client.buffersProcessed.Load()),
			zap.Uint64("buffersDropped", client.buffersDropped.Load()),
		)
		buf, bufferReadSuccess := msg.(*buffer.Buffer)
		if bufferReadSuccess {
			// sleep until retry time
			for client.RetryAfter().After(time.Now()) {
				client.sleep(client.RetryAfter(), buf)
			}
			if !buf.HasEvents() {
				client.Logger.Warn("Buffer is empty, skipping", buf.ZapStats()...)
				client.buffersProcessed.Add(1)
				continue
			}
			if client.sendBufferWithRetryPolicy(buf) {
				client.buffersProcessed.Add(1)
			}
		} else {
			client.Logger.Error(
				"Cannot convert message to Buffer",
				zap.String("session", session),
				zap.Any("msg", msg),
			)
			client.buffersBroken.Add(1)
		}
		client.lastAcceptedAt.Store(time.Now().UnixNano())
	}
}

// Sends buffer to DataSet. If not succeeds and try is possible (it retryable), try retry until possible (timeout)
func (client *DataSetClient) sendBufferWithRetryPolicy(buf *buffer.Buffer) bool {
	// Do not use NewExponentialBackOff since it calls Reset and the code here must
	// call Reset after changing the InitialInterval (this saves an unnecessary call to Now).
	expBackoff := backoff.ExponentialBackOff{
		InitialInterval:     client.Config.BufferSettings.RetryInitialInterval,
		RandomizationFactor: client.Config.BufferSettings.RetryRandomizationFactor,
		Multiplier:          client.Config.BufferSettings.RetryMultiplier,
		MaxInterval:         client.Config.BufferSettings.RetryMaxInterval,
		MaxElapsedTime:      client.Config.BufferSettings.RetryMaxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	expBackoff.Reset()
	retryNum := int64(0)
	for {
		response, payloadLen, err := client.sendAddEventsBuffer(buf)
		client.setLastError(err)
		lastHttpStatus := uint32(0)
		if err != nil {
			client.Logger.Error("unable to send addEvents buffers", zap.Error(err))
			client.setLastErrorTimestamp(time.Now())
			if !strings.Contains(err.Error(), "Unable to send request") {
				lastHttpStatus = HttpErrorHasErrorMessage
				client.LastHttpStatus.Store(lastHttpStatus)
				client.onBufferDrop(buf, lastHttpStatus, err)
				return false // exit loop (failed to send buffer)
			}
			lastHttpStatus = HttpErrorCannotConnect
			client.LastHttpStatus.Store(lastHttpStatus)
		}
		zaps := make([]zap.Field, 0)
		if response.ResponseObj != nil {
			zaps = append(
				zaps,
				zap.String("httpStatus", response.ResponseObj.Status),
				zap.Int("httpCode", response.ResponseObj.StatusCode),
			)
			lastHttpStatus = uint32(response.ResponseObj.StatusCode)
			client.LastHttpStatus.Store(lastHttpStatus)
		}

		zaps = append(
			zaps,
			zap.String("status", response.Status),
			zap.String("message", response.Message),
		)
		client.Logger.Debug("Events were sent to DataSet",
			zaps...,
		)

		if isOkStatus(lastHttpStatus) {
			// everything was fine, there is no need for retries
			client.bytesAPIAccepted.Add(uint64(payloadLen))
			return true // exit loop (buffer sent)
		}

		backoffDelay := expBackoff.NextBackOff()
		client.Logger.Info("Backoff + Retries: ", zap.Int64("retryNum", retryNum), zap.Duration("backoffDelay", backoffDelay))
		if backoffDelay == backoff.Stop {
			// throw away the batch
			err = fmt.Errorf("max elapsed time expired %w", err)
			client.onBufferDrop(buf, lastHttpStatus, err)
			return false // exit loop (failed to send buffer)
		}

		if isRetryableStatus(lastHttpStatus) {
			// check whether header is specified and get its value
			retryAfter, specified := client.getRetryAfter(
				response.ResponseObj,
				backoffDelay,
			)

			if specified {
				// retry after is specified, we should update
				// client state, so we do not send more requests

				client.setRetryAfter(retryAfter)
			}

			client.sleep(retryAfter, buf)
			buf.SetStatus(buffer.Retrying)
		} else {
			err = fmt.Errorf("non recoverable error %w", err)
			client.onBufferDrop(buf, lastHttpStatus, err)
			return false // exit loop (failed to send buffer)
		}
		retryNum++
	}
}

func (client *DataSetClient) statisticsSweeper() {
	for i := uint64(0); ; i++ {
		client.logStatistics()
		// wait for some time before new sweep
		time.Sleep(15 * time.Second)
	}
}

// Statistics returns statistics about events, buffers processing from the start time
func (client *DataSetClient) Statistics() *Statistics {
	// for how long are events being processed
	firstAt := time.Unix(0, client.firstReceivedAt.Load())
	lastAt := time.Unix(0, client.lastAcceptedAt.Load())
	processingDur := lastAt.Sub(firstAt)
	processingInSec := processingDur.Seconds()

	// if nothing was processed, do not log statistics
	if processingInSec <= 0 {
		return nil
	}

	// log buffer stats
	bProcessed := client.buffersProcessed.Load()
	bEnqueued := client.buffersEnqueued.Load()
	bDropped := client.buffersDropped.Load()
	bBroken := client.buffersBroken.Load()

	buffersStats := QueueStats{
		bEnqueued,
		bProcessed,
		bDropped,
		bBroken,
		processingDur,
	}

	// log events stats
	eProcessed := client.eventsProcessed.Load()
	eEnqueued := client.eventsEnqueued.Load()
	eDropped := client.eventsDropped.Load()
	eBroken := client.eventsBroken.Load()

	eventsStats := QueueStats{
		eEnqueued,
		eProcessed,
		eDropped,
		eBroken,
		processingDur,
	}

	// log transferred stats
	bAPISent := client.bytesAPISent.Load()
	bAPIAccepted := client.bytesAPIAccepted.Load()
	transferStats := TransferStats{
		bAPISent,
		bAPIAccepted,
		bProcessed,
		processingDur,
	}

	return &Statistics{
		Buffers:  buffersStats,
		Events:   eventsStats,
		Transfer: transferStats,
	}
}

func (client *DataSetClient) logStatistics() {
	stats := client.Statistics()
	if stats == nil {
		return
	}

	b := stats.Buffers
	client.Logger.Info(
		"Buffers' Queue Stats:",
		zap.Uint64("processed", b.Processed()),
		zap.Uint64("enqueued", b.Enqueued()),
		zap.Uint64("dropped", b.Dropped()),
		zap.Uint64("broken", b.Broken()),
		zap.Uint64("waiting", b.Waiting()),
		zap.Float64("successRate", b.SuccessRate()),
		zap.Float64("processingS", b.ProcessingTime().Seconds()),
		zap.Duration("processing", b.ProcessingTime()),
	)

	// log events stats
	e := stats.Events
	client.Logger.Info(
		"Events' Queue Stats:",
		zap.Uint64("processed", e.Processed()),
		zap.Uint64("enqueued", e.Enqueued()),
		zap.Uint64("dropped", e.Dropped()),
		zap.Uint64("broken", e.Broken()),
		zap.Uint64("waiting", e.Waiting()),
		zap.Float64("successRate", e.SuccessRate()),
		zap.Float64("processingS", e.ProcessingTime().Seconds()),
		zap.Duration("processing", e.ProcessingTime()),
	)

	// log transferred stats
	mb := float64(1024 * 1024)
	t := stats.Transfer
	client.Logger.Info(
		"Transfer Stats:",
		zap.Float64("bytesSentMB", float64(t.BytesSent())/mb),
		zap.Float64("bytesAcceptedMB", float64(t.BytesAccepted())/mb),
		zap.Float64("throughputMBpS", t.ThroughputBpS()/mb),
		zap.Uint64("buffersProcessed", t.BuffersProcessed()),
		zap.Float64("perBufferMB", t.AvgBufferBytes()/mb),
		zap.Float64("successRate", t.SuccessRate()),
		zap.Float64("processingS", t.ProcessingTime().Seconds()),
		zap.Duration("processing", t.ProcessingTime()),
	)
}

func (client *DataSetClient) bufferSweeper(lifetime time.Duration) {
	client.Logger.Info("Starting buffer sweeper with lifetime", zap.Duration("lifetime", lifetime))
	totalKept := atomic.Uint64{}
	totalSwept := atomic.Uint64{}
	for i := uint64(0); ; i++ {
		// if everything was finished, there is no need to run buffer sweeper
		//if client.finished.Load() {
		//	client.Logger.Info("Stopping buffer sweeper", zap.Uint64("sweepId", i))
		//	break
		//}
		kept := atomic.Uint64{}
		swept := atomic.Uint64{}
		client.Logger.Debug("Buffer sweeping started", zap.Uint64("sweepId", i))
		buffers := client.getBuffers()
		for _, buf := range buffers {
			// publish buffers that are ready only
			// if we are actively adding events into this buffer skip it for now
			if buf.ShouldSendAge(lifetime) {
				if buf.HasStatus(buffer.Ready) {
					client.publishBuffer(buf)
					swept.Add(1)
					totalSwept.Add(1)
				} else {
					buf.PublishAsap.Store(true)
					kept.Add(1)
					totalKept.Add(1)
				}
			} else {
				kept.Add(1)
				totalKept.Add(1)
			}
		}

		// log just every n-th sweep
		lvl := zap.DebugLevel
		if i%100 == 0 {
			lvl = zap.InfoLevel
		}
		client.Logger.Log(
			lvl,
			"Buffer sweeping finished",
			zap.Uint64("sweepId", i),
			zap.Uint64("nowKept", kept.Load()),
			zap.Uint64("nowSwept", swept.Load()),
			zap.Uint64("nowCombined", kept.Load()+swept.Load()),
			zap.Uint64("totalKept", totalKept.Load()),
			zap.Uint64("totalSwept", totalSwept.Load()),
			zap.Uint64("totalCombined", totalKept.Load()+totalSwept.Load()),
		)

		time.Sleep(lifetime)
	}
}

func (client *DataSetClient) publishBuffer(buf *buffer.Buffer) {
	if buf.HasStatus(buffer.Publishing) {
		// buffer is already publishing, this should not happen
		// so lets skip it
		client.Logger.Warn("Buffer is already beeing published", buf.ZapStats()...)
		return
	}

	// we are manipulating with client.buffer, so lets lock it
	client.buffersAllMutex.Lock()
	originalStatus := buf.Status()
	buf.SetStatus(buffer.Publishing)

	// if the buffer is being used, lets create new one as replacement
	if originalStatus.IsActive() {
		newBuf := buffer.NewEmptyBuffer(buf.Session, buf.Token)
		client.Logger.Debug(
			"Removing buffer for session",
			zap.String("session", buf.Session),
			zap.String("oldUuid", buf.Id.String()),
			zap.String("newUuid", newBuf.Id.String()),
		)

		client.initBuffer(newBuf, buf.SessionInfo())
		client.buffers[buf.Session] = newBuf
	}
	client.buffersAllMutex.Unlock()

	client.Logger.Debug("publishing buffer", buf.ZapStats()...)

	// publish buffer so it can be sent
	client.buffersEnqueued.Add(+1)
	client.BufferPerSessionTopic.Pub(buf, buf.Session)
}

// Exporter rejects handling of incoming batches if is in error state
func (client *DataSetClient) isInErrorState() (bool, error) {
	// In case one of session failed (with retryable status) to send request (batch of event to DataSet), client retries sending this request. Unless retry attempts timed out
	if isRetryableStatus(client.LastHttpStatus.Load()) && !client.isLastRetryableErrorTimedOut() {
		err := client.LastError()
		if err != nil {
			return true, fmt.Errorf("rejecting - Last HTTP request contains an error: %w", err)
		} else {
			return true, fmt.Errorf("rejecting - Last HTTP request had retryable status")
		}
	}

	// DataSet can limit rate using RetryAfter header. In such case client rejects incoming events
	if client.RetryAfter().After(time.Now()) {
		return true, fmt.Errorf("rejecting - should retry after %s", client.RetryAfter().Format(time.RFC1123))
	}

	return false, nil
}

func (client *DataSetClient) getRetryAfter(response *http.Response, def time.Duration) (time.Time, bool) {
	if response == nil {
		return time.Now().Add(def), false
	}
	after := response.Header.Get("Retry-After")

	// if it's not specified, return default
	if len(after) == 0 {
		return time.Now().Add(def), false
	}

	// it is specified in seconds?
	afterS, errS := strconv.Atoi(after)
	if errS == nil {
		// it's integer => so lets convert it
		return time.Now().Add(time.Duration(afterS) * time.Second), true
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Date
	afterT, errT := time.Parse(time.RFC1123, after)
	if errT == nil {
		return afterT, true
	}

	client.Logger.Warn("Illegal value of Retry-After, using default",
		zap.String("retryAfter", after),
		zap.Duration("default", def),
		zap.Error(errS),
		zap.Error(errT),
	)

	return time.Now().Add(def), false
}

// TODO move this function outside of DataSetClient to not misunderstand that shared client is sleeping. It each go routine which is sleeping.
func (client *DataSetClient) sleep(retryAfter time.Time, buffer *buffer.Buffer) {
	// truncate current time to wait a little bit longer
	sleepFor := retryAfter.Sub(time.Now().Truncate(time.Second))
	client.Logger.Info(
		"RetryAfter is in the future, waiting...",
		zap.Duration("sleepFor", sleepFor),
		zap.String("bufferSession", buffer.Session),
		zap.String("bufferUuid", buffer.Id.String()),
	)
	time.Sleep(sleepFor)
}

func (client *DataSetClient) LastError() error {
	client.lastErrorMu.RLock()
	defer client.lastErrorMu.RUnlock()
	return client.lastError
}

func (client *DataSetClient) setLastError(err error) {
	client.lastErrorMu.Lock()
	defer client.lastErrorMu.Unlock()
	client.lastError = err
}

func (client *DataSetClient) RetryAfter() time.Time {
	client.retryAfterMu.RLock()
	defer client.retryAfterMu.RUnlock()
	return client.retryAfter
}

func (client *DataSetClient) setRetryAfter(t time.Time) {
	client.retryAfterMu.Lock()
	defer client.retryAfterMu.Unlock()
	client.retryAfter = t
}

func (client *DataSetClient) onBufferDrop(buf *buffer.Buffer, status uint32, err error) {
	client.buffersDropped.Add(1)
	client.Logger.Error("Dropping buffer",
		buf.ZapStats(
			zap.Uint32("httpStatus", status),
			zap.Error(err),
		)...,
	)
}

func (client *DataSetClient) setLastErrorTimestamp(timestamp time.Time) {
	client.lastErrorMu.Lock()
	defer client.lastErrorMu.Unlock()
	client.lastErrorTs = timestamp
}

func (client *DataSetClient) lastErrorTimestamp() time.Time {
	client.retryAfterMu.RLock()
	defer client.retryAfterMu.RUnlock()
	return client.lastErrorTs
}

func (client *DataSetClient) isLastRetryableErrorTimedOut() bool {
	return client.lastErrorTimestamp().Add(client.Config.BufferSettings.RetryMaxElapsedTime).Before(time.Now())
}

func (client *DataSetClient) ServerHost() string {
	return client.serverHost
}
