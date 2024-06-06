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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/scalyr/dataset-go/pkg/api/response"

	"github.com/scalyr/dataset-go/pkg/api/add_events"
	"github.com/scalyr/dataset-go/pkg/api/request"
	"github.com/scalyr/dataset-go/pkg/buffer"

	"go.uber.org/zap"
)

const (
	errMsgUnableToSentRequest   = "unable to send request"
	errMsgUnableToReadResponse  = "unable to read response"
	errMsgUnableToParseResponse = "unable to parse response"
)

/*
Wrapper around: https://app.scalyr.com/help/api#addEvents
*/

type EventWithMeta struct {
	EventBundle *add_events.EventBundle
	Key         string
	SessionInfo add_events.SessionInfo
}

func NewEventWithMeta(
	bundle *add_events.EventBundle,
	groupBy []string,
	serverHost string,
	debug bool,
) EventWithMeta {
	// initialise
	key := ""
	info := make(add_events.SessionInfo)

	// adjust server host attribute
	adjustServerHost(bundle, serverHost)

	// construct Key
	bundleKey := extractKeyAndUpdateInfo(bundle, groupBy, key, info)

	// in debug mode include bundleKey
	if debug {
		info[add_events.AttrSessionKey] = bundleKey
	}

	return EventWithMeta{
		EventBundle: bundle,
		Key:         bundleKey,
		SessionInfo: info,
	}
}

func extractKeyAndUpdateInfo(bundle *add_events.EventBundle, groupBy []string, key string, info add_events.SessionInfo) string {
	// construct key and move attributes from attrs to sessionInfo
	for _, k := range groupBy {
		val, ok := bundle.Event.Attrs[k]
		key += k + ":"
		if ok {
			key += fmt.Sprintf("%s", val)

			// move to session info and remove from attributes
			info[k] = val
			delete(bundle.Event.Attrs, k)
		}
		key += "___DELIM___"
	}

	// use md5 to shorten the key
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func adjustServerHost(bundle *add_events.EventBundle, serverHost string) {
	// figure out serverHost value
	usedServerHost := serverHost
	// if event's ServerHost is set, use it
	if len(bundle.Event.ServerHost) > 0 {
		usedServerHost = bundle.Event.ServerHost
	} else {
		// if somebody is using library directly and forget to set Event.ServerHost,
		// lets check the attributes first

		// check serverHost attribute
		attrHost, ok := bundle.Event.Attrs[add_events.AttrServerHost]
		if ok {
			usedServerHost = attrHost.(string)
		} else {
			// if serverHost attribute is not set, check the __origServerHost
			attrOrigHost, okOrig := bundle.Event.Attrs[add_events.AttrOrigServerHost]
			if okOrig {
				usedServerHost = attrOrigHost.(string)
			}
		}
	}

	// for the SessionInfo, it has to be in the serverHost
	// therefore remove the orig and set the serverHost
	// so that it can be in the next step moved to sessionInfo
	delete(bundle.Event.Attrs, add_events.AttrOrigServerHost)
	bundle.Event.Attrs[add_events.AttrServerHost] = usedServerHost
}

// AddEvents enqueues given events for processing (sending to Dataset).
// It returns an error if the batch was not accepted (e.g. exporter in error state and retrying handle previous batches or client is being shutdown).
func (client *DataSetClient) AddEvents(bundles []*add_events.EventBundle) error {
	if client.finished.Load() {
		return fmt.Errorf("client has finished - rejecting all new events")
	}
	shouldReject, errR := client.isInErrorState()
	if shouldReject {
		return fmt.Errorf("AddEvents - reject batch: %w", errR)
	}

	// then, figure out which keys are part of the batch
	// store there information about the host
	bundlesWithMeta := make(map[string][]EventWithMeta)
	for _, bundle := range bundles {
		bWM := NewEventWithMeta(bundle, client.Config.BufferSettings.GroupBy, client.serverHost, client.Config.Debug)

		session := fmt.Sprintf("%s-%s", client.Id, bWM.Key)
		list, found := bundlesWithMeta[session]
		if !found {
			bundlesWithMeta[session] = []EventWithMeta{bWM}
		} else {
			bundlesWithMeta[session] = append(list, bWM)
		}
	}

	// update time when the first batch was received
	if client.firstReceivedAt.Load() == 0 {
		client.firstReceivedAt.Store(time.Now().UnixNano())
	}

	// then create all subscribers
	// add subscriber for events by key
	// add subscriber for buffer by key
	client.addEventsMutex.Lock()
	for key, list := range bundlesWithMeta {
		_, found := client.eventBundleSubscriptionChannels[key]
		if !found {
			// add information about the host to the sessionInfo
			client.newBufferForEvents(key, &list[0].SessionInfo)

			client.newEventBundleSubscriberRoutine(key)
		}
	}
	client.addEventsMutex.Unlock()

	// and as last step - publish them

	for key, list := range bundlesWithMeta {
		for _, bundle := range list {
			client.eventBundlePerKeyTopic.Pub(bundle, key)
			client.statistics.EventsEnqueuedAdd(1)
		}
	}

	return nil
}

func (client *DataSetClient) newEventBundleSubscriberRoutine(key string) {
	ch := client.eventBundlePerKeyTopic.Sub(key)
	client.eventBundleSubscriptionChannels[key] = ch
	go (func(session string, ch chan interface{}) {
		client.listenAndSendBundlesForKey(key, ch)
	})(key, ch)
}

func (client *DataSetClient) newBufferForEvents(session string, info *add_events.SessionInfo) {
	buf := buffer.NewEmptyBuffer(session, client.Config.Tokens.WriteLog)

	client.initBuffer(buf, info)

	client.buffersAllMutex.Lock()
	client.buffers[session] = buf
	defer client.buffersAllMutex.Unlock()

	// create subscriber, so all the upcoming buffers are processed as well
	client.newBuffersSubscriberRoutine(session)
}

func (client *DataSetClient) listenAndSendBundlesForKey(key string, ch chan interface{}) {
	client.Logger.Debug("Listening to events with key",
		zap.String("key", key),
	)

	// this function has to be called from AddEvents - inner loop
	// it assumes that all bundles have the same key
	getBuffer := func(key string) *buffer.Buffer {
		buf := client.getBuffer(key)
		// change state to mark that bundles are being added
		buf.SetStatus(buffer.AddingBundles)
		return buf
	}

	publish := func(key string, buf *buffer.Buffer) *buffer.Buffer {
		client.publishBuffer(buf)
		return getBuffer(key)
	}

	for processedMsgCnt := 0; ; processedMsgCnt++ {
		msg, channelReceiveSuccess := <-ch
		if !channelReceiveSuccess {
			client.Logger.Error(
				"Cannot receive EventWithMeta from channel",
				zap.String("key", key),
				zap.Any("msg", msg),
			)
			client.statistics.EventsBrokenAdd(1)
			client.lastAcceptedAt.Store(time.Now().UnixNano())
			continue
		}

		bundle, ok := msg.(EventWithMeta)
		if ok {
			buf := getBuffer(key)

			added, err := buf.AddBundle(bundle.EventBundle)
			if err != nil {
				if errors.Is(err, &buffer.NotAcceptingError{}) {
					buf = getBuffer(key)
				} else {
					client.Logger.Error("Cannot add bundle", zap.Error(err))
					// TODO: what to do?
					// this error happens when the buffer is already publishing
					// we can solve it by waiting little bit
					// since we are also returning TooMuch we will get new buffer
					// few lines below
					time.Sleep(10 * time.Millisecond)
				}
			}

			if buf.ShouldSendSize() || added == buffer.TooMuch && buf.HasEvents() {
				buf = publish(key, buf)
			}

			if added == buffer.TooMuch {
				added, err = buf.AddBundle(bundle.EventBundle)
				if err != nil {
					if errors.Is(err, &buffer.NotAcceptingError{}) {
						buf = getBuffer(key)
					} else {
						client.Logger.Error("Cannot add bundle", zap.Error(err))
						client.statistics.EventsDroppedAdd(1)
						continue
					}
				}
				if buf.ShouldSendSize() {
					buf = publish(key, buf)
				}
				if added == buffer.TooMuch {
					client.Logger.Fatal("Bundle was not added for second time!", buf.ZapStats()...)
					client.statistics.EventsDroppedAdd(1)
					continue
				}
			}
			client.statistics.EventsProcessedAdd(1)

			buf.SetStatus(buffer.Ready)
			// it could happen that the buffer could have been published
			// by buffer sweeper, but it was skipped, because we have been
			// adding events, so lets check it and publish it if needed
			if buf.PublishAsap.Load() {
				client.publishBuffer(buf)
			}
		} else {
			_, purgeReadSuccess := msg.(Purge)
			if purgeReadSuccess {
				break
			}
			client.Logger.Error(
				"Cannot convert message to EventWithMeta",
				zap.String("key", key),
				zap.Any("msg", msg),
			)
			client.statistics.EventsBrokenAdd(1)
		}
	}
}

// isProcessingBuffers returns True if there are still some unprocessed buffers.
// False otherwise.
func (client *DataSetClient) isProcessingBuffers() bool {
	return client.statistics.BuffersEnqueued() > (client.statistics.BuffersProcessed() + client.statistics.BuffersDropped() + client.statistics.BuffersBroken())
}

// isProcessingEvents returns True if there are still some unprocessed events.
// False otherwise.
func (client *DataSetClient) isProcessingEvents() bool {
	return client.statistics.EventsEnqueued() > (client.statistics.EventsProcessed() + client.statistics.EventsDropped() + client.statistics.EventsBroken())
}

// Shutdown takes care of shutdown of client. It does following steps
// - stops processing of new events,
// - tries (with 1st half of shutdownMaxTimeout period) to process (add into buffers) all the events,
// - tries (with 2nd half of shutdownMaxTimeout period) to send processed events (buffers) into DataSet
func (client *DataSetClient) Shutdown() error {
	client.Logger.Info("Shutting down - BEGIN")
	// start measuring processing time
	processingStart := time.Now()

	// mark as finished to prevent processing of further events
	client.finished.Store(true)

	// log statistics when finish was called
	client.logStatistics()

	retryShutdownTimeout := client.Config.BufferSettings.RetryShutdownTimeout
	maxElapsedTime := retryShutdownTimeout/2 - 100*time.Millisecond
	client.Logger.Info(
		"Shutting down - waiting for events",
		zap.Duration("maxElapsedTime", maxElapsedTime),
		zap.Duration("retryShutdownTimeout", retryShutdownTimeout),
		zap.Duration("elapsedTime", time.Since(processingStart)),
	)

	var lastError error = nil
	expBackoff := backoff.ExponentialBackOff{
		InitialInterval:     client.Config.BufferSettings.RetryInitialInterval,
		RandomizationFactor: client.Config.BufferSettings.RetryRandomizationFactor,
		Multiplier:          client.Config.BufferSettings.RetryMultiplier,
		MaxInterval:         client.Config.BufferSettings.RetryMaxInterval,
		MaxElapsedTime:      maxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}

	// try (with timeout) to process (add into buffers) events,
	retryNum := 0
	expBackoff.Reset()
	initialEventsDropped := client.statistics.EventsDropped()
	for client.isProcessingEvents() {
		// log statistics
		client.logStatistics()

		backoffDelay := expBackoff.NextBackOff()
		client.Logger.Info(
			"Shutting down - processing events",
			zap.Int("retryNum", retryNum),
			zap.Duration("backoffDelay", backoffDelay),
			zap.Uint64("eventsEnqueued", client.statistics.EventsEnqueued()),
			zap.Uint64("eventsProcessed", client.statistics.EventsProcessed()),
			zap.Duration("elapsedTime", time.Since(processingStart)),
			zap.Duration("maxElapsedTime", maxElapsedTime),
		)
		if backoffDelay == expBackoff.Stop {
			break
		}
		time.Sleep(backoffDelay)
		retryNum++
	}

	// send all buffers
	client.Logger.Info(
		"Shutting down - publishing all buffers",
		zap.Duration("retryShutdownTimeout", retryShutdownTimeout),
		zap.Duration("elapsedTime", time.Since(processingStart)),
	)
	client.publishAllBuffers()

	// reinitialize expBackoff with MaxElapsedTime based on actually elapsed time of processing (previous phase)
	processingElapsed := time.Since(processingStart)
	maxElapsedTime = maxDuration(retryShutdownTimeout-processingElapsed, retryShutdownTimeout/2)
	client.Logger.Info(
		"Shutting down - waiting for buffers",
		zap.Duration("maxElapsedTime", maxElapsedTime),
		zap.Duration("retryShutdownTimeout", retryShutdownTimeout),
		zap.Duration("elapsedTime", time.Since(processingStart)),
	)

	expBackoff = backoff.ExponentialBackOff{
		InitialInterval:     client.Config.BufferSettings.RetryInitialInterval,
		RandomizationFactor: client.Config.BufferSettings.RetryRandomizationFactor,
		Multiplier:          client.Config.BufferSettings.RetryMultiplier,
		MaxInterval:         client.Config.BufferSettings.RetryMaxInterval,
		MaxElapsedTime:      maxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	// do wait (with timeout) for all buffers to be sent to the server
	retryNum = 0
	expBackoff.Reset()
	initialBuffersDropped := client.statistics.BuffersDropped()
	for client.isProcessingBuffers() {
		// log statistics
		client.logStatistics()

		backoffDelay := expBackoff.NextBackOff()
		client.Logger.Info(
			"Shutting down - processing buffers",
			zap.Int("retryNum", retryNum),
			zap.Duration("backoffDelay", backoffDelay),
			zap.Uint64("buffersEnqueued", client.statistics.BuffersEnqueued()),
			zap.Uint64("buffersProcessed", client.statistics.BuffersProcessed()),
			zap.Uint64("buffersDropped", client.statistics.BuffersDropped()),
			zap.Duration("elapsedTime", time.Since(processingStart)),
			zap.Duration("maxElapsedTime", maxElapsedTime),
		)
		if backoffDelay == expBackoff.Stop {
			break
		}
		time.Sleep(backoffDelay)
		retryNum++
	}

	// construct error messages
	if client.isProcessingEvents() {
		lastError = fmt.Errorf(
			"not all events have been processed - %d",
			client.statistics.EventsEnqueued()-client.statistics.EventsProcessed(),
		)
		client.Logger.Error(
			"Shutting down - not all events have been processed",
			zap.Uint64("eventsEnqueued", client.statistics.EventsEnqueued()),
			zap.Uint64("eventsProcessed", client.statistics.EventsProcessed()),
		)
	}

	if client.isProcessingBuffers() {
		lastError = fmt.Errorf(
			"not all buffers have been processed - %d",
			client.statistics.BuffersEnqueued()-client.statistics.BuffersProcessed()-client.statistics.BuffersDropped(),
		)
		client.Logger.Error(
			"Shutting down - not all buffers have been processed",
			zap.Int("retryNum", retryNum),
			zap.Uint64("buffersEnqueued", client.statistics.BuffersEnqueued()),
			zap.Uint64("buffersProcessed", client.statistics.BuffersProcessed()),
			zap.Uint64("buffersDropped", client.statistics.BuffersDropped()),
		)
	}

	eventsDropped := client.statistics.EventsDropped() - initialEventsDropped
	if eventsDropped > 0 {
		lastError = fmt.Errorf(
			"some events were dropped during finishing - %d",
			eventsDropped,
		)
		client.Logger.Error(
			"Shutting down - events were dropped during shutdown",
			zap.Uint64("eventsDropped", eventsDropped),
		)
	}

	buffersDropped := client.statistics.BuffersDropped() - initialBuffersDropped
	if buffersDropped > 0 {
		lastError = fmt.Errorf(
			"some buffers were dropped during finishing - %d",
			buffersDropped,
		)
		client.Logger.Error(
			"Shutting down - buffers were dropped during shutdown",
			zap.Uint64("buffersDropped", buffersDropped),
		)
	}

	// print final statistics
	client.logStatistics()

	if lastError == nil {
		client.Logger.Info(
			"Shutting down - success",
			zap.Duration("retryShutdownTimeout", retryShutdownTimeout),
			zap.Duration("elapsedTime", time.Since(processingStart)),
		)
	} else {
		client.Logger.Error(
			"Shutting down - error", zap.Error(lastError),
			zap.Duration("retryShutdownTimeout", retryShutdownTimeout),
			zap.Duration("elapsedTime", time.Since(processingStart)),
		)
		if client.LastError() == nil {
			return lastError
		}
	}

	return client.LastError()
}

func (client *DataSetClient) sendAddEventsBuffer(buf *buffer.Buffer) (*add_events.AddEventsResponse, int, error) {
	client.Logger.Debug("Sending buf", buf.ZapStats()...)

	payload, err := buf.Payload()
	if err != nil {
		client.Logger.Warn("Cannot create payload", buf.ZapStats(zap.Error(err))...)
		return nil, 0, fmt.Errorf("cannot create payload: %w", err)
	}
	client.Logger.Debug("Created payload",
		buf.ZapStats(
			zap.Int("payload", len(payload)),
			zap.Float64("payloadRatio", float64(len(payload))/buffer.ShouldSentBufferSize),
		)...,
	)
	resp := &add_events.AddEventsResponse{}
	client.statistics.PayloadSizeRecord(int64(len(payload)))

	httpRequest, err := request.NewApiRequest(
		http.MethodPost, client.addEventsEndpointUrl,
	).WithWriteLog(client.Config.Tokens).WithPayload(payload).WithUserAgent(client.userAgent).HttpRequest()
	if err != nil {
		return nil, len(payload), fmt.Errorf("cannot create request: %w", err)
	}

	apiCallStart := time.Now()
	err = client.apiCall(httpRequest, resp)
	apiCallEnd := time.Now()
	client.statistics.ResponseTimeRecord(apiCallEnd.Sub(apiCallStart))
	client.statistics.BytesAPISentAdd(uint64(len(payload)))

	if strings.HasPrefix(resp.Status, "error") {
		client.Logger.Error(
			"Problematic payload",
			zap.String("message", resp.Message),
			zap.String("status", resp.Status),
			zap.Int("payloadLength", len(payload)),
		)
	}

	return resp, len(payload), err
}

//func (client *DataSetClient) groupBundles(bundles []*add_events.EventBundle) map[string][]*add_events.EventBundle {
//	grouped := make(map[string][]*add_events.EventBundle)
//
//	// group batch
//	for _, bundle := range bundles {
//		if bundle == nil {
//			continue
//		}
//		key := bundle.Key(client.Config.GroupBy)
//		grouped[key] = append(grouped[key], bundle)
//	}
//	client.Logger.Debug("Batch was grouped",
//		zap.Int("batchSize", len(bundles)),
//		zap.Int("distinctStreams", len(grouped)),
//	)
//	return grouped
//}

func (client *DataSetClient) apiCall(req *http.Request, response response.ResponseObjSetter) error {
	resp, err := client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", errMsgUnableToSentRequest, err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			client.Logger.Error("Error when closing:", zap.Error(err))
		}
	}()

	client.Logger.Debug("Received response",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.Int64("contentLength", resp.ContentLength),
	)

	if resp.StatusCode != http.StatusOK {
		// TODO improve docs - why dont we return error in this case
		client.Logger.Warn(
			"!!!!! PAYLOAD WAS NOT ACCEPTED !!!!",
			zap.Int("statusCode", resp.StatusCode),
		)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", errMsgUnableToReadResponse, err)
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return fmt.Errorf("%s: %w, url: %s, status: %d, response: %s", errMsgUnableToParseResponse, err, client.addEventsEndpointUrl, resp.StatusCode, truncateText(string(responseBody), 1000))
	}

	response.SetResponseObj(resp)

	return nil
}

// publishAllBuffers send all buffers to the server
func (client *DataSetClient) publishAllBuffers() {
	buffers := client.getBuffers()
	client.Logger.Info(
		"Publish all buffers",
		zap.Int("numBuffers", len(buffers)),
	)
	for _, buf := range buffers {
		client.publishBuffer(buf)
	}
}

func (client *DataSetClient) getBuffers() []*buffer.Buffer {
	buffers := make([]*buffer.Buffer, 0)
	client.buffersAllMutex.Lock()
	defer client.buffersAllMutex.Unlock()
	for _, buf := range client.buffers {
		buffers = append(buffers, buf)
	}
	return buffers
}

// Truncate provided text to the provided length
func truncateText(text string, length int) string {
	if len(text) > length {
		text = string([]byte(text)[:length]) + "..."
	}

	return text
}

func maxDuration(a, b time.Duration) time.Duration {
	if a >= b {
		return a
	}
	return b
}
