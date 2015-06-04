package i

type Combat interface {
	UUID() string                   // The UUID.
	CommandQueue() chan interface{} // The command queue.
	WhileLocked(func())             // Lock then do something.
}
