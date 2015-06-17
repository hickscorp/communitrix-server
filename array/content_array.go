package array

import (
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
)

type ContentArray struct {
	Size    *logic.Vector
	Content [][][]int
}

// Instanciator.
func NewContentArray(size *logic.Vector, filler ContentArrayFiller) *ContentArray {
	if filler == nil {
		filler = NewIntContentArrayFiller(0)
	}
	ret := ContentArray{Size: size.Clone()}
	ret.Content = make([][][]int, size.X)
	at := &logic.Vector{0, 0, 0}
	for at.X = range ret.Content {
		yRow := make([][]int, size.Y)
		ret.Content[at.X] = yRow
		for at.Y = range yRow {
			zRow := make([]int, size.Z)
			yRow[at.Y] = zRow
			for at.Z = range zRow {
				zRow[at.Z] = filler(at)
			}
		}
	}
	return &ret
}

func NewContentArrayFromPiece(piece *logic.Piece, filler ContentArrayFiller) *ContentArray {
	return (&ContentArray{Size: piece.Size.Clone()}).FromPiece(piece, filler)
}
func (this *ContentArray) FromPiece(piece *logic.Piece, filler ContentArrayFiller) *ContentArray {
	if filler == nil {
		filler = NewIntContentArrayFiller(0)
	}
	this.Size.FromVector(piece.Size)
	this.Content = make([][][]int, this.Size.X)
	at := &logic.Vector{0, 0, 0}
	for at.X = range this.Content {
		yRow := make([][]int, this.Size.Y)
		this.Content[at.X] = yRow
		for at.Y = range yRow {
			zRow := make([]int, this.Size.Z)
			yRow[at.Y] = zRow
			for at.Z = range zRow {
				zRow[at.Z] = filler(at)
			}
		}
	}
	for _, cell := range piece.Content {
		this.Content[cell.X][cell.Y][cell.Z] = cell.Value
	}
	return this
}

// Clone allows to deep-copy an array.
func (this *ContentArray) Clone() *ContentArray {
	return NewContentArray(this.Size, NewCopyContentArrayFiller(this))
}

// ToPiece converts an array to a piece object.
func (this *ContentArray) ToPiece() *logic.Piece {
	ret := logic.NewPiece(this.Size, 0)
	this.Each(func(at *logic.Vector, val int) {
		ret.AddCell(logic.NewCellFromValues(at.X, at.Y, at.Z, val))
	})
	return ret
}

// Each allows to perform a given function over each of this object's components. The current object is then returned for chaining.
func (this *ContentArray) Each(do func(*logic.Vector, int)) {
	at := logic.Vector{0, 0, 0}
	for at.X = 0; at.X < this.Size.X; at.X++ {
		for at.Y = 0; at.Y < this.Size.Y; at.Y++ {
			for at.Z = 0; at.Z < this.Size.Z; at.Z++ {
				do(&at, this.Content[at.X][at.Y][at.Z])
			}
		}
	}
}
