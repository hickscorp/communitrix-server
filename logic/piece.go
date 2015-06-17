package logic

import (
	"github.com/op/go-logging"
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

func (this *Piece) CleanUp() *Piece {
	log.Debug("Piece cleanup: Size: %d, Cells: %d.", this.Size, len(this.Content))
	// No cells in piece? BUG.
	if this.IsEmpty() {
		log.Warning("A piece was asked to clean itself up, but it does not contain any cell.")
		return this
	}

	// Compute local limits.
	xMin, yMin, zMin, xMax, yMax, zMax := 1000.0, 1000.0, 1000.0, -1000.0, -1000.0, -1000.0
	for _, cell := range this.Content {
		x, y, z := float64(cell.X), float64(cell.Y), float64(cell.Z)
		xMin, yMin, zMin = math.Min(xMin, x), math.Min(yMin, y), math.Min(zMin, z)
		xMax, yMax, zMax = math.Max(xMax, x), math.Max(yMax, y), math.Max(zMax, z)
	}
	// Make a vector for minimum limits.
	min := NewVectorFromValues(util.QuickIntRound(xMin), util.QuickIntRound(yMin), util.QuickIntRound(zMin))
	log.Debug("  - Computed Min: %d", min)
	// Make another vector for maximum limuts.
	max := NewVectorFromValues(util.QuickIntRound(xMax), util.QuickIntRound(yMax), util.QuickIntRound(zMax))
	log.Debug("  - Computed Max: %d", max)
	// Set the new piece size.
	this.Size = min.Clone().Inv().Translate(max).Translate(NewVectorFromValues(1, 1, 1))
	log.Debug("  - Computed Size: %d", this.Size)
	// Pinpoint the center of the piece, invert it.
	halved := min.Clone().Inv().Translate(max).Clone().Half().Inv()
	log.Debug("  - Will apply translation: %d", halved)
	// Translate all this piece's cells.
	this.Translate(halved)
	return this
}

// Clone allows to deep-copy a piece.
func (this *Piece) Clone() *Piece {
	return &Piece{
		Size:    this.Size.Clone(),
		Content: this.Content.Clone(),
	}
}

// Translate applies a given translation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Translate(t *Vector) *Piece {
	for _, v := range this.Content {
		v.Translate(t)
	}
	return this
}

// Rotate applies a given rotation transformation to the current object. The current object is then returned for chaining.
func (this *Piece) Rotate(q *Quaternion) *Piece {
	this.Size.Rotate(q)
	for _, v := range this.Content {
		v.Rotate(q)
	}
	return this
}

func (this *Piece) IsEmpty() bool {
	return len(this.Content) == 0
}

func (this *Piece) AddCell(cell *Cell) *Piece {
	this.Content = append(this.Content, cell)
	return this
}
