package logic

type Cells []*Cell

func (this Cells) Clone() Cells {
	ret := make(Cells, len(this))
	for i, cell := range this {
		ret[i] = cell.Clone()
	}
	return ret
}
