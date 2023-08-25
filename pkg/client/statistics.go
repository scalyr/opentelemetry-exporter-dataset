package client

import (
	"time"
)

// QueueStats stores statistics related to the queue processing
type QueueStats struct {
	// enqueued is number of items that has been accepted for processing
	enqueued uint64 `mapstructure:"enqueued"`
	// Processed is number of items that has been successfully processed
	processed uint64 `mapstructure:"processed"`
	// dropped is number of items that has been dropped since they couldn't be processed
	dropped uint64 `mapstructure:"dropped"`
	// broken is number of items that has been damaged by queue, should be zero
	broken uint64 `mapstructure:"broken"`
	// processingTime is duration of the processing
	processingTime time.Duration `mapstructure:"processingTime"`
}

// Enqueued is number of items that has been accepted for processing
func (stats QueueStats) Enqueued() uint64 {
	return stats.enqueued
}

// Processed is number of items that has been successfully processed
func (stats QueueStats) Processed() uint64 {
	return stats.processed
}

// Dropped is number of items that has been dropped since they couldn't be processed
func (stats QueueStats) Dropped() uint64 {
	return stats.dropped
}

// Broken is number of items that has been damaged by queue, should be zero
func (stats QueueStats) Broken() uint64 {
	return stats.broken
}

// Waiting is number of items that are waiting for being processed
// Enqueued - Processed - Dropped - Broken
func (stats QueueStats) Waiting() uint64 {
	return stats.enqueued - stats.processed - stats.dropped - stats.broken
}

// SuccessRate of items processing
// (Processed - Dropped - Broken) / Processed
func (stats QueueStats) SuccessRate() float64 {
	if stats.processed > 0 {
		return float64(stats.processed-stats.dropped-stats.broken) / float64(stats.processed)
	} else {
		return 0.0
	}
}

// ProcessingTime is duration of the processing
func (stats QueueStats) ProcessingTime() time.Duration {
	return stats.processingTime
}

// TransferStats stores statistics related to the data transfers
type TransferStats struct {
	// bytesSent is the amount of bytes that were sent to the server
	// each retry is counted
	bytesSent uint64 `mapstructure:"bytesSentMB"`
	// bytesAccepted is the amount of MB that were accepted by the server
	// retries are not counted
	bytesAccepted uint64 `mapstructure:"bytesAcceptedMB"`
	// buffersProcessed is number of processed buffers
	buffersProcessed uint64
	// processingTime is duration of the processing
	processingTime time.Duration `mapstructure:"processingTime"`
}

// BytesSent is the amount of bytes that were sent to the server
// each retry is counted
func (stats TransferStats) BytesSent() uint64 {
	return stats.bytesSent
}

// BytesAccepted is the amount of MB that were accepted by the server
// retries are not counted
func (stats TransferStats) BytesAccepted() uint64 {
	return stats.bytesAccepted
}

// BuffersProcessed is number of processed buffers
func (stats TransferStats) BuffersProcessed() uint64 {
	return stats.buffersProcessed
}

// ThroughputBpS is the throughput based on BytesAccepted
func (stats TransferStats) ThroughputBpS() float64 {
	return float64(stats.bytesAccepted) / stats.processingTime.Seconds()
}

// SuccessRate of the transfer - BytesAccepted / BytesSent
func (stats TransferStats) SuccessRate() float64 {
	if stats.bytesSent > 0 {
		return float64(stats.bytesAccepted) / float64(stats.bytesSent)
	} else {
		return 0.0
	}
}

// AvgBufferBytes is average buffer size in bytes - BytesAccepted / BuffersProcessed
func (stats TransferStats) AvgBufferBytes() float64 {
	if stats.buffersProcessed > 0 {
		return float64(stats.bytesAccepted) / float64(stats.buffersProcessed)
	} else {
		return 0
	}
}

// ProcessingTime is duration of the processing
func (stats TransferStats) ProcessingTime() time.Duration {
	return stats.processingTime
}

// Statistics store statistics about queues and transferred data
// These are statistics from the beginning of the processing
type Statistics struct {
	// Events stores statistics about processing events
	Events QueueStats `mapstructure:"events"`
	// Buffers stores statistics about processing buffers
	Buffers QueueStats `mapstructure:"buffers"`
	// Transfer stores statistics about data transfers
	Transfer TransferStats `mapstructure:"transfer"`
}
