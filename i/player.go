package i

import (
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"net"
)

type Player interface {
	UUID() string                     // UUID.
	Username() string                 // Username.
	SetUsername(username string)      // Setter on Username.
	Level() int                       // Level.
	Connection() net.Conn             // Connection.
	Combat() Combat                   // Combat if any.
	CommandQueue() chan<- interface{} // Command queue, write-only to the outside world.
	AsSendable() util.MapHelper       // Serialization.
	WhileLocked(do func())            // Lock then do something.
	IsInCombat() bool                 // Query combat.
	JoinCombat(combat Combat)         // Join a combat.
	LeaveCombat() bool                // Leave a combat.
}
