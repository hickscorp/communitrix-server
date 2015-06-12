package array

import (
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
)

type ContentArray struct {
	Size    *logic.Vector
	Content [][][]int
}

// Define our array filling template method.
type ContentArrayFiller func(at *logic.Vector) int

// NewIntContentArrayFiller creates a new routine which will fill an array with a constant value.
func NewIntContentArrayFiller(val int) ContentArrayFiller {
	return func(*logic.Vector) int { return val }
}

// NewCopyContentArrayFiller will copy values from a given tri-ensional array.
func NewCopyContentArrayFiller(other *ContentArray) ContentArrayFiller {
	return func(at *logic.Vector) int { return other.Content[at.X][at.Y][at.Z] }
}

// Instanciator.
func NewContentArray(size *logic.Vector, filler ContentArrayFiller) *ContentArray {
	if filler == nil {
		filler = NewIntContentArrayFiller(0)
	}
	ret := ContentArray{Size: size}
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
func NewContentArrayFromPiece(piece *logic.Piece) *ContentArray {
	ret := NewContentArray(piece.Size, nil)
	for _, cell := range piece.Content {
		ret.Content[cell.X][cell.Y][cell.Z] = cell.Value
	}
	return ret
}

// Clone allows to deep-copy an array.
func (this *ContentArray) Clone() *ContentArray {
	return NewContentArray(this.Size, NewCopyContentArrayFiller(this))
}

func (this *ContentArray) ToPiece() *logic.Piece {
	ret := logic.NewPiece(this.Size, 0)
	this.Each(func(at *logic.Vector, val int) {
		ret.Content = append(ret.Content, logic.NewCellFromInts(at.X, at.Y, at.Z, val))
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
