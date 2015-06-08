package i

import "gogs.pierreqr.fr/doodloo/communitrix/util"

type Combat interface {
	UUID() string                     // The UUID.
	CommandQueue() chan<- interface{} // The command queue, write-only to the outside world.
	AsSendable() util.MapHelper       // Serialization.
	WhileLocked(func())               // Lock then do something.
}
