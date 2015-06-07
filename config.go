package main

import "github.com/op/go-logging"

// Config is the main configuration object.
type Config struct {
	Port                 *int
	HubCommandBufferSize *int
	ClientSendBufferSize *int
	Seed                 *int64
	LogLevel             logging.Level
}
