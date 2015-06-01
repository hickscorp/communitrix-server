package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"net"
	"os"
	"time"
)

var config = Config{}
var log = logging.MustGetLogger("communitrix")
var format = logging.MustStringFormatter("%{color}%{level:.1s} %{shortfunc}%{color:reset} %{message}")

func init() {
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "communitrix")
}

func main() {
	// Allows to parse a single parameter, the port.
	config.Port = flag.Int("port", 8080, "Port to serve on.")
	config.HubCommandBufferSize = flag.Int("hubCommandBuffer", 4096, "Size of the hub command queue buffer.")
	config.ClientSendBufferSize = flag.Int("clientSendBufferSize", 64, "Size of the client send queue buffer.")
	config.MaximumMessageSize = flag.Int64("maximumMessageSize", 32768, "Maximum message size allowed, expressed in bytes.")
	config.PongTimeout = flag.Duration("pongTimeout", 10*time.Second, "Maximum muted time allowed for a client, expressed in seconds.")
	config.AutosaveInterval = flag.Duration("autosaveInterval", 15*time.Minute, "Interval for auto-saving channels data.")
	logLevel := flag.String("logLevel", "WARNING", "Log level [DEBUG|INFO|WARNING|ERROR|CRITICAL].")
	flag.Parse()

	config.LogLevel, _ = logging.LogLevel(*logLevel)
	logging.SetLevel(config.LogLevel, "communitrix")

	// Prepare our listen address.
	addr := fmt.Sprintf("0.0.0.0:%d", *config.Port)
	// Listen for incoming connections.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer listener.Close()
	// cmd := WrapCommand(NCmdWelcome{Message: "Hi there!"})

	// Create and run our hub.
	hub := RunNewHub()

	fmt.Println("Listening on " + addr)
	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go hub.HandleClient(conn)
	}
}
