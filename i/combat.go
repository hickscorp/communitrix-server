package i

type Combat interface {
	UUID() string                     // The UUID.
	CommandQueue() chan<- interface{} // The command queue, write-only to the outside world.
	WhileLocked(func())               // Lock then do something.
}
