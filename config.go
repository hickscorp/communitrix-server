package main

import (
	"github.com/op/go-logging"
	"time"
)

// Config is the main configuration object.
type Config struct {
	Port                 *int
	HubCommandBufferSize *int
	ClientSendBufferSize *int
	MaximumMessageSize   *int64
	PongTimeout          *time.Duration
	AutosaveInterval     *time.Duration
	LogLevel             logging.Level
}
