package logic

type ContentArray struct {
	Size    *Vector
	Content [][][]int
}

func NewContentArray(size *Vector, filler ArrayFiller) ContentArray {
	if filler == nil {
		filler = NewIntArrayFiller(0)
	}
	ret := ContentArray{Size: size}
	ret.Content = make([][][]int, size.X)
	at := &Vector{0, 0, 0}
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
	return ret
}

func (this ContentArray) Clone() ContentArray {
	return NewContentArray(this.Size, NewCopyArrayFiller(this))
}
