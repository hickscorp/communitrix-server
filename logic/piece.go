package logic

import (
	"gogs.pierreqr.fr/doodloo/communitrix/i"
)

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

// Clone allows to deep-copy a piece.
func (this Piece) Clone() *Piece {
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
