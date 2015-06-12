package gen

import (
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math"
)

type PieceSplitter struct {
}

func NewPieceSplitter() *PieceSplitter {
	return &PieceSplitter{}
}

func (this *PieceSplitter) Run(originalPiece *logic.Piece, count int) (logic.Pieces, bool) {
	// params validation
	if originalPiece == nil {
		log.Warning("PieceSplitter cannot run without a piece.")
		return nil, false
	} else if count <= 0 {
		log.Warning("PieceSplitter break down a piece into a negative count of pieces.")
		return nil, false
	} else if count >= len(originalPiece.Content) {
		log.Warning("PieceSplitter cannot break down a piece to more than its count of cells.")
		return nil, false
	}

	ret := make(logic.Pieces, 0)
	pieces := originalPiece.Content.Clone()
	originalPiece.Content = make(logic.Cells, 0)
	for i := 0; len(pieces) > 5; i++ {
		var xMin, yMin, zMin = 1000.0, 1000.0, 1000.0
		for _, cell := range pieces {
			xMin, yMin, zMin = math.Min(float64(cell.X), xMin), math.Min(float64(cell.Y), yMin), math.Min(float64(cell.Z), zMin)
		}
		min := logic.NewVectorFromInts(util.QuickIntRound(xMin), util.QuickIntRound(yMin), util.QuickIntRound(zMin))

		nextCells := make(logic.Cells, 0)
		cutCells := make(logic.Cells, 0)
		xMinNew, yMinNew, zMinNew, xMaxNew, yMaxNew, zMaxNew := 1000.0, 1000.0, 1000.0, -1000.0, -1000.0, -1000.0
		// Left cut.
		if i%3 == 0 {
			for _, cell := range pieces {
				if cell.X <= min.X+1 {
					cell.Value = i + 1
					cutCells = append(cutCells, cell)
					xMinNew, yMinNew, zMinNew = math.Min(float64(cell.X), xMinNew), math.Min(float64(cell.Y), yMinNew), math.Min(float64(cell.Z), zMinNew)
					xMaxNew, yMaxNew, zMaxNew = math.Max(float64(cell.X), xMaxNew), math.Max(float64(cell.Y), yMaxNew), math.Max(float64(cell.Z), zMaxNew)
					originalPiece.Content = append(originalPiece.Content, cell)
				} else {
					nextCells = append(nextCells, cell)
				}
			}
			// Bottom cut.
		} else if i%3 == 1 {
			for _, cell := range pieces {
				if cell.Y <= min.Y+1 {
					cell.Value = i + 1
					cutCells = append(cutCells, cell)
					xMinNew, yMinNew, zMinNew = math.Min(float64(cell.X), xMinNew), math.Min(float64(cell.Y), yMinNew), math.Min(float64(cell.Z), zMinNew)
					xMaxNew, yMaxNew, zMaxNew = math.Max(float64(cell.X), xMaxNew), math.Max(float64(cell.Y), yMaxNew), math.Max(float64(cell.Z), zMaxNew)
					originalPiece.Content = append(originalPiece.Content, cell)
				} else {
					nextCells = append(nextCells, cell)
				}
			}
			// Forward cut.
		} else if i%3 == 2 {
			for _, cell := range pieces {
				if cell.Z <= min.Z+1 {
					cell.Value = i + 1
					cutCells = append(cutCells, cell)
					xMinNew, yMinNew, zMinNew = math.Min(float64(cell.X), xMinNew), math.Min(float64(cell.Y), yMinNew), math.Min(float64(cell.Z), zMinNew)
					xMaxNew, yMaxNew, zMaxNew = math.Max(float64(cell.X), xMaxNew), math.Max(float64(cell.Y), yMaxNew), math.Max(float64(cell.Z), zMaxNew)
					originalPiece.Content = append(originalPiece.Content, cell)
				} else {
					nextCells = append(nextCells, cell)
				}
			}
		}

		minNew := logic.NewVectorFromInts(util.QuickIntRound(xMinNew), util.QuickIntRound(yMinNew), util.QuickIntRound(zMinNew))
		maxNew := logic.NewVectorFromInts(util.QuickIntRound(xMaxNew), util.QuickIntRound(yMaxNew), util.QuickIntRound(zMaxNew))
		delNew := minNew.Clone()
		delNew.Inv()
		delNew.Translate(maxNew)
		delNew.Abs()

		newPiece := logic.NewPiece(delNew, 0)
		newPiece.Content = cutCells
		ret = append(ret, newPiece)

		pieces = nextCells
	}

	return ret, true
}
