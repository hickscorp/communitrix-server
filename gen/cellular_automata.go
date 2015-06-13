package gen

import (
	"gogs.pierreqr.fr/doodloo/communitrix/array"
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math/rand"
)

type CellularAutomata struct {
	size            *logic.Vector
	result          *array.ContentArray
	probabilities   *array.ContentArray
	spreadingFactor float64
}

func NewCellularAutomata(size *logic.Vector) *CellularAutomata {
	return &CellularAutomata{
		size:            size.Clone(),
		spreadingFactor: 0.1,
	}
}

// Run creates the unit.
func (this *CellularAutomata) Run(density float64) (*logic.Piece, bool) {
	// Normalize inputs.
	if (density < 0.0 || density > 1.0) ||
		(this.size.X <= 0 || this.size.X&1 == 0) ||
		(this.size.Y <= 0 || this.size.Y&1 == 0) ||
		(this.size.Z <= 0 || this.size.Z&1 == 0) {
		return nil, false
	}
	// Prepare the total number of blocks to be created.
	targetSize := int(float64(this.size.Volume()) * density)
	// Prepare the result and probabilities array.
	this.probabilities, this.result = array.NewContentArray(this.size, nil), array.NewContentArray(this.size, nil)
	// Cache the shape center.
	center := this.size.Clone()
	center.Half()

	// Set the first block inside the results array.
	this.fillCell(center, -1)
	// Keep track of the total number of cells we've added.
	totalCellsAdded := 1

	for {
		// Compute the target cells to generate during this iteration.
		cellsPerIteration := util.QuickIntRound(float64(totalCellsAdded) * this.spreadingFactor)
		if cellsPerIteration == 0 {
			cellsPerIteration = 1
		} else if cellsPerIteration > 100 {
			cellsPerIteration = 100
		}
		// Whenever we're reaching the target number of cells, cap it correctly.
		if totalCellsAdded+cellsPerIteration > targetSize {
			cellsPerIteration = targetSize - totalCellsAdded
		}

		// Prepare iteration.
		probSum := 0
		groups := map[int]logic.Vectors{}
		// Look at every single array element.
		this.probabilities.Each(func(at *logic.Vector, pro int) {
			if pro != 0 {
				if locations, ok := groups[pro]; ok {
					groups[pro] = append(locations, at.Clone())
				} else {
					probSum += pro
					groups[pro] = logic.Vectors{at.Clone()}
				}
			}
		})
		// Build a dice array.
		dice := make([]int, 0, probSum)
		for pro := range groups {
			for c := 0; c < pro; c++ {
				dice = append(dice, pro)
			}
		}
		for i := 0; i < cellsPerIteration; i++ {
			pro := dice[rand.Intn(probSum)]
			locations := groups[pro]
			location := locations[rand.Intn(len(locations))]
			if !this.fillCell(location, -1) {
				i--
			}
		}
		// Whenever we reach the target size, stop.
		totalCellsAdded += cellsPerIteration
		if totalCellsAdded >= targetSize {
			break
		}
	}

	// Split into pieces.
	this.split(5)

	// Generate the piece.
	piece := logic.NewPiece(this.size, totalCellsAdded-1)
	this.result.Each(func(at *logic.Vector, val int) {
		if val != 0 {
			piece.AddCell(logic.NewCellFromVectorAndValue(at, val))
		}
	})
	this.result, this.probabilities = nil, nil
	return piece, true
}

func (this *CellularAutomata) fillCell(at *logic.Vector, val int) bool {
	if this.result.Content[at.X][at.Y][at.Z] != 0 {
		return false
	}
	this.probabilities.Content[at.X][at.Y][at.Z] = 0
	this.result.Content[at.X][at.Y][at.Z] = val
	for _, d := range directions {
		pos := at.Clone().Translate(d)
		if pos.X >= 0 && pos.X < this.size.X && pos.Y >= 0 && pos.Y < this.size.Y && pos.Z >= 0 && pos.Z < this.size.Z {
			// If there is already a probability there, increase it.
			actual := this.probabilities.Content[pos.X][pos.Y][pos.Z]
			this.probabilities.Content[pos.X][pos.Y][pos.Z] = actual + 1
		}
	}
	return true
}

func (this *CellularAutomata) split(count int) {
}
