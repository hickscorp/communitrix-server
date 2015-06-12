package gen

import "gogs.pierreqr.fr/doodloo/communitrix/logic"

// Our direction checks.
var directions = logic.Vectors{
	&logic.Vector{-1, 0, 0}, &logic.Vector{+1, 0, 0}, // Left / Right.
	&logic.Vector{0, -1, 0}, &logic.Vector{0, +1, 0}, // Top / Bottom.
	&logic.Vector{0, 0, -1}, &logic.Vector{0, 0, +1}, // Forward / Backward
}
