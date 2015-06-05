package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"net"
	"os"
)

var (
	config = Config{}
	log    = logging.MustGetLogger("communitrix")
	format = logging.MustStringFormatter("%{color}%{level:.1s} %{shortfunc}%{color:reset} %{message}")
)

func init() {
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "communitrix")
}

func main() {
	// Allows to parse a single parameter, the port.
	config.Port = flag.Int("port", 9003, "Port to serve on.")
	config.HubCommandBufferSize = flag.Int("hubCommandBuffer", 4096, "Size of the hub command queue buffer.")
	config.ClientSendBufferSize = flag.Int("clientSendBufferSize", 64, "Size of the client send queue buffer.")
	logLevel := flag.String("logLevel", "WARNING", "Log level [DEBUG|INFO|WARNING|ERROR|CRITICAL].")
	flag.Parse()

	config.LogLevel, _ = logging.LogLevel(*logLevel)
	logging.SetLevel(config.LogLevel, "communitrix")

	// Prepare our listen address.
	addr := fmt.Sprintf("0.0.0.0:%d", *config.Port)
	// Listen for incoming connections.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("Error listening: %s", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer listener.Close()
	log.Info("Server is ready on %s.", addr)
	// Create and run our hub.
	hub := NewHub()
	go hub.Run()
	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting new client: %s", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go hub.HandleClient(conn)
	}
}
