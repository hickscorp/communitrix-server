package array

import "gogs.pierreqr.fr/doodloo/communitrix/logic"

type ContentArray struct {
	Size    *logic.Vector
	Content [][][]int
}

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

func (this ContentArray) Clone() *ContentArray {
	return NewContentArray(this.Size, NewCopyContentArrayFiller(this))
}
