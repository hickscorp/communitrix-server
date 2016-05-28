package gen

import (
	"github.com/op/go-logging"
	"github.com/hickscorp/communitrix-server/logic"
)

var (
	// Our logger.
	log = logging.MustGetLogger("communitrix")
	// Our direction checks.
	directions = logic.Vectors{
		&logic.Vector{-1, 0, 0}, &logic.Vector{+1, 0, 0}, // Left / Right.
		&logic.Vector{0, -1, 0}, &logic.Vector{0, +1, 0}, // Top / Bottom.
		&logic.Vector{0, 0, -1}, &logic.Vector{0, 0, +1}, // Forward / Backward
	}
)
