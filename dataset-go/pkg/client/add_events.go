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

/*
Wrapper around: https://app.scalyr.com/help/api#addEvents
*/

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
	seenKeys := make(map[string]bool)
	for _, bundle := range bundles {
		key := bundle.Key(client.Config.BufferSettings.GroupBy)
		seenKeys[key] = true
	}

	// update time when the first batch was received
	if client.firstReceivedAt.Load() == 0 {
		client.firstReceivedAt.Store(time.Now().UnixNano())
	}

	// then create all subscribers
	// add subscriber for events by key
	// add subscriber for buffer by key
	client.addEventsMutex.Lock()
	defer client.addEventsMutex.Unlock()
	for key := range seenKeys {
		_, found := client.eventBundleSubscriptionChannels[key]
		if !found {
			// add information about the host to the sessionInfo
			client.newBufferForEvents(key)

			client.newEventBundleSubscriberRoutine(key)
		}
	}

	// and as last step - publish them
	for _, bundle := range bundles {
		key := bundle.Key(client.Config.BufferSettings.GroupBy)
		client.eventBundlePerKeyTopic.Pub(bundle, key)
		client.eventsEnqueued.Add(1)
	}

	return nil
}

// fixServerHostsInBundle fills the attribute __origServerHost for the event
// and removes the attribute serverHost. This is needed to properly associate
// incoming events with the correct host
func (client *DataSetClient) fixServerHostsInBundle(bundle *add_events.EventBundle) {
	delete(bundle.Event.Attrs, add_events.AttrServerHost)

	// set the attribute __origServerHost to the event's ServerHost
	if len(bundle.Event.ServerHost) > 0 {
		bundle.Event.Attrs[add_events.AttrOrigServerHost] = bundle.Event.ServerHost
		return
	}

	// as fallback use the value set to the client
	bundle.Event.Attrs[add_events.AttrOrigServerHost] = client.serverHost
}

func (client *DataSetClient) newEventBundleSubscriberRoutine(key string) {
	ch := client.eventBundlePerKeyTopic.Sub(key)
	client.eventBundleSubscriptionChannels[key] = ch
	go (func(session string, ch chan interface{}) {
		client.listenAndSendBundlesForKey(key, ch)
	})(key, ch)
}

func (client *DataSetClient) newBufferForEvents(key string) {
	session := fmt.Sprintf("%s-%s", client.Id, key)
	buf := buffer.NewEmptyBuffer(session, client.Config.Tokens.WriteLog)

	client.initBuffer(buf, client.SessionInfo)

	client.buffersAllMutex.Lock()
	client.buffers[session] = buf
	defer client.buffersAllMutex.Unlock()

	// create subscriber, so all the upcoming buffers are processed as well
	client.newBuffersSubscriberRoutine(session)
}

func (client *DataSetClient) listenAndSendBundlesForKey(key string, ch chan interface{}) {
	client.Logger.Info("Listening to events with key",
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
				"Cannot receive EventBundle from channel",
				zap.String("key", key),
				zap.Any("msg", msg),
			)
			client.eventsBroken.Add(1)
			client.lastAcceptedAt.Store(time.Now().UnixNano())
			continue
		}

		bundle, ok := msg.(*add_events.EventBundle)
		if ok {
			buf := getBuffer(key)
			client.fixServerHostsInBundle(bundle)

			added, err := buf.AddBundle(bundle)
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
				added, err = buf.AddBundle(bundle)
				if err != nil {
					if errors.Is(err, &buffer.NotAcceptingError{}) {
						buf = getBuffer(key)
					} else {
						client.Logger.Error("Cannot add bundle", zap.Error(err))
						client.eventsDropped.Add(1)
						continue
					}
				}
				if buf.ShouldSendSize() {
					buf = publish(key, buf)
				}
				if added == buffer.TooMuch {
					client.Logger.Fatal("Bundle was not added for second time!", buf.ZapStats()...)
					client.eventsDropped.Add(1)
					continue
				}
			}
			client.eventsProcessed.Add(1)

			buf.SetStatus(buffer.Ready)
			// it could happen that the buffer could have been published
			// by buffer sweeper, but it was skipped, because we have been
			// adding events, so lets check it and publish it if needed
			if buf.PublishAsap.Load() {
				client.publishBuffer(buf)
			}
		} else {
			client.Logger.Error(
				"Cannot convert message to EventBundle",
				zap.String("key", key),
				zap.Any("msg", msg),
			)
			client.eventsBroken.Add(1)
		}
	}
}

// isProcessingBuffers returns True if there are still some unprocessed buffers.
// False otherwise.
func (client *DataSetClient) isProcessingBuffers() bool {
	return client.buffersEnqueued.Load() > (client.buffersProcessed.Load() + client.buffersDropped.Load() + client.buffersBroken.Load())
}

// isProcessingEvents returns True if there are still some unprocessed events.
// False otherwise.
func (client *DataSetClient) isProcessingEvents() bool {
	return client.eventsEnqueued.Load() > (client.eventsProcessed.Load() + client.eventsDropped.Load() + client.eventsBroken.Load())
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
	initialEventsDropped := client.eventsDropped.Load()
	for client.isProcessingEvents() {
		// log statistics
		client.logStatistics()

		backoffDelay := expBackoff.NextBackOff()
		client.Logger.Info(
			"Shutting down - processing events",
			zap.Int("retryNum", retryNum),
			zap.Duration("backoffDelay", backoffDelay),
			zap.Uint64("eventsEnqueued", client.eventsEnqueued.Load()),
			zap.Uint64("eventsProcessed", client.eventsProcessed.Load()),
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
	initialBuffersDropped := client.buffersDropped.Load()
	for client.isProcessingBuffers() {
		// log statistics
		client.logStatistics()

		backoffDelay := expBackoff.NextBackOff()
		client.Logger.Info(
			"Shutting down - processing buffers",
			zap.Int("retryNum", retryNum),
			zap.Duration("backoffDelay", backoffDelay),
			zap.Uint64("buffersEnqueued", client.buffersEnqueued.Load()),
			zap.Uint64("buffersProcessed", client.buffersProcessed.Load()),
			zap.Uint64("buffersDropped", client.buffersDropped.Load()),
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
			client.eventsEnqueued.Load()-client.eventsProcessed.Load(),
		)
		client.Logger.Error(
			"Shutting down - not all events have been processed",
			zap.Uint64("eventsEnqueued", client.eventsEnqueued.Load()),
			zap.Uint64("eventsProcessed", client.eventsProcessed.Load()),
		)
	}

	if client.isProcessingBuffers() {
		lastError = fmt.Errorf(
			"not all buffers have been processed - %d",
			client.buffersEnqueued.Load()-client.buffersProcessed.Load()-client.buffersDropped.Load(),
		)
		client.Logger.Error(
			"Shutting down - not all buffers have been processed",
			zap.Int("retryNum", retryNum),
			zap.Uint64("buffersEnqueued", client.buffersEnqueued.Load()),
			zap.Uint64("buffersProcessed", client.buffersProcessed.Load()),
			zap.Uint64("buffersDropped", client.buffersDropped.Load()),
		)
	}

	eventsDropped := client.eventsDropped.Load() - initialEventsDropped
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

	buffersDropped := client.buffersDropped.Load() - initialBuffersDropped
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

	httpRequest, err := request.NewApiRequest(
		http.MethodPost, client.addEventsEndpointUrl,
	).WithWriteLog(client.Config.Tokens).WithPayload(payload).WithUserAgent(client.userAgent).HttpRequest()
	if err != nil {
		return nil, len(payload), fmt.Errorf("cannot create request: %w", err)
	}

	err = client.apiCall(httpRequest, resp)
	client.bytesAPISent.Add(uint64(len(payload)))

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
		return fmt.Errorf("unable to send request: %w", err)
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
		return fmt.Errorf("unable to read response: %w", err)
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return fmt.Errorf("unable to parse response body: %w, url: %s, response: %s", err, client.addEventsEndpointUrl, truncateText(string(responseBody), 1000))
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
