package gen

import "gogs.pierreqr.fr/doodloo/communitrix/array"

type Generator interface {
	Run() array.ContentArrayFiller
}
