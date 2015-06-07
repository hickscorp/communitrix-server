package logic

type TriDimArray [][][]int

func NewTriDimArray(size *Vector) TriDimArray {
	return NewFilledTriDimArray(size, nil)
}
func NewFilledTriDimArray(size *Vector, filler ArrayFiller) TriDimArray {
	ret := make(TriDimArray, size.X)
	at := &Vector{0, 0, 0}
	for at.X = range ret {
		yRow := make([][]int, size.Y)
		ret[at.X] = yRow
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

func (this TriDimArray) Size(sizeOrNil *Vector) *Vector {
	if sizeOrNil == nil {
		sizeOrNil = &Vector{0, 0, 0}
	}
	if sizeOrNil.X = len(this); sizeOrNil.X > 0 {
		if sizeOrNil.Y = len(this[0]); sizeOrNil.Y > 0 {
			sizeOrNil.Z = len(this[0][0])
		}
	}
	return sizeOrNil
}
func (this TriDimArray) Clone() TriDimArray {
	return NewFilledTriDimArray(this.Size(&Vector{0, 0, 0}), NewCopyArrayFiller(this))
}
