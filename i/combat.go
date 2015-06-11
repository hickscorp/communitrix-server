package i

import (
	//"gogs.pierreqr.fr/doodloo/communitrix/cmd/cbt"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
)

type Combat interface {
	UUID() string               // The UUID.
	Notify(interface{})         // Send something to the combat.
	AsSendable() util.MapHelper // Serialization.
	WhileLocked(func())         // Lock then do something.
	Run()                       // Start running the combat events loop.
}
