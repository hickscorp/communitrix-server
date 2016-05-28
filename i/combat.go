package i

import (
	"github.com/hickscorp/communitrix-server/util"
)

type Combat interface {
	UUID() string                  // The UUID.
	Notify(interface{})            // Send something to the combat.
	AsSendable() util.MapHelper    // Serialization.
	Run()                          // Start running the combat events loop.
	Summarize(chan util.MapHelper) // Gets the summary of a combat.
}
