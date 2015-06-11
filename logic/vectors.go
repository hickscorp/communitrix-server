package logic

type Vectors []*Vector

func (this Vectors) Clone() Vectors {
	ret := make(Vectors, len(this))
	for i, cell := range this {
		ret[i] = cell.Clone()
	}
	return ret
}
