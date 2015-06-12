package logic

import (
	"github.com/op/go-logging"
	"gogs.pierreqr.fr/doodloo/communitrix/i"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math"
)

var log = logging.MustGetLogger("communitrix")

type Piece struct {
	Size    *Vector `json:"size"`
	Content Cells   `json:"content"`
}

// Instanciator.
func NewPiece(size *Vector, capacity int) *Piece {
	return &Piece{
		Size:    size.Clone(),
		Content: make(Cells, 0, capacity),
	}
}

func (this *Piece) CleanUp() {
	xMin, yMin, zMin, xMax, yMax, zMax := 1000.0, 1000.0, 1000.0, -1000.0, -1000.0, -1000.0
	for _, cell := range this.Content {
		x, y, z := float64(cell.X), float64(cell.Y), float64(cell.Z)
		xMin, yMin, zMin = math.Min(xMin, x), math.Min(yMin, y), math.Min(zMin, z)
		xMax, yMax, zMax = math.Max(xMax, x), math.Max(yMax, y), math.Max(zMax, z)
	}
	min := NewVectorFromInts(util.QuickIntRound(xMin), util.QuickIntRound(yMin), util.QuickIntRound(zMin))
	max := NewVectorFromInts(util.QuickIntRound(xMax), util.QuickIntRound(yMax), util.QuickIntRound(zMax))

	this.Size = min.Clone()
	this.Size.Inv()
	this.Size.Translate(max)
	halved := this.Size.Clone()
	halved.Half()
	halved.Inv()
	this.Size.Translate(NewVectorFromInts(1, 1, 1))
	this.Translate(halved)
}

// Clone allows to deep-copy a piece.
func (this *Piece) Clone() *Piece {
	return &Piece{
		Size:    this.Size.Clone(),
		Content: this.Content.Clone(),
	}
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Translate(t i.Localizable) {
	for _, v := range this.Content {
		v.Translate(t)
	}
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Rotate(q *Quaternion) {
	this.Size.Rotate(q)
	for _, v := range this.Content {
		v.Rotate(q)
	}
}
