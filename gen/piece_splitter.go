package gen

import (
	"github.com/hickscorp/communitrix-server/array"
	"github.com/hickscorp/communitrix-server/logic"
	"math/rand"
	"sync"
)

type PieceSplitter struct {
}

func NewPieceSplitter() *PieceSplitter {
	return &PieceSplitter{}
}

func (this *PieceSplitter) Run(piece *logic.Piece, count int) (logic.Pieces, bool) {
	// params validation
	if piece == nil {
		log.Warning("PieceSplitter cannot run without a piece.")
		return nil, false
	} else if count <= 0 {
		log.Warning("PieceSplitter break down a piece into a negative count of pieces.")
		return nil, false
	} else if count >= len(piece.Content) {
		log.Warning("PieceSplitter cannot break down a piece to more than its count of cells.")
		return nil, false
	}

	// Convert the piece to an array.
	arr := array.NewContentArrayFromPiece(piece, nil)
	// Prepare our synchronization objects.
	wg := &sync.WaitGroup{}
	wg.Add(1)

	ready, queryable := make(chan bool), make(chan *availabilityQuery, count)
	// Prepare our pieces array to return.
	pieces := make(logic.Pieces, count)

	// Start as many recursors as required.
	log.Debug("Spawning %d recursors.", count)
	for i := 0; i < count; i++ {
		piece = logic.NewPiece(logic.NewVectorFromValues(0, 0, 0), 0)
		pieces[i] = piece
		go RunNewRecursor(ready, wg, i+1, arr.Size, queryable)
	}

	// This will hold the number of recursors that are ready.
	allReady := false
	readyCount := 0
	done := 0
	var onNilSignal func(q *availabilityQuery)
	// When all our recursors become ready, this closure will handle nil signals.
	onNilWhenReady := func(q *availabilityQuery) {
		done++
		log.Debug("  - Recursor #%d ( %d / %d ) is done.", q.ID, done, count)
	}
	// As long as our recursors are not all ready, this closure will handle nil signals.
	onNilWhenNotReady := func(q *availabilityQuery) {
		if readyCount++; readyCount == count {
			allReady = true
			onNilSignal = onNilWhenReady
			wg.Done()
		}
	}
	onNilSignal = onNilWhenNotReady

	// Now that we have everything ready and running, start the synching mechanism.
	for done = 0; done < count; {
		select {
		// A recursor is asking for availability of a cell.
		case q := <-queryable:
			// A recursor gave us a nil signal.
			if q.At == nil {
				onNilSignal(q)
				continue
			}
			free := arr.Content[q.At.X][q.At.Y][q.At.Z] == -1
			if free {
				arr.Content[q.At.X][q.At.Y][q.At.Z] = q.ID
				pieces[q.ID-1].AddCell(logic.NewCellFromValues(q.At.X, q.At.Y, q.At.Z, q.ID))
			}
			q.Answer <- free
		}
	}
	return pieces.CleanUp(), true
}

type availabilityQuery struct {
	ID     int
	Answer chan bool
	At     *logic.Vector
}

func RunNewRecursor(ready chan bool, start *sync.WaitGroup, id int, bounds *logic.Vector, queryChan chan *availabilityQuery) {
	// Prepare a unique query object for this recursor.
	query := &availabilityQuery{id, make(chan bool), nil}
	// Prepare a function for signaling a nil event. First one means ready, second one means we're done.
	signalNil := func() {
		query.At = nil
		queryChan <- query
	}
	// Get rid of the answer channel whenever this recursor exits. Also signal we're done.
	defer func() {
		close(query.Answer)
		signalNil()
	}()

	// Create a querying closure.
	queryFree := func(at *logic.Vector) bool {
		// No need to look in the content array if we're already out of bounds.
		if at.X < 0 || at.X >= bounds.X || at.Y < 0 || at.Y >= bounds.Y || at.Z < 0 || at.Z >= bounds.Z {
			return false
		}
		// Put the query position in the query object, and send it.
		query.At = at
		queryChan <- query
		// Wait for the answer, return it.
		return <-query.Answer
	}

	// Create our recursive closure.
	var recurse func(*logic.Vector)
	recurse = func(at *logic.Vector) {
		dir := directions.Clone().Shuffle()
		pos := logic.NewVectorFromValues(0, 0, 0)
		for _, d := range dir {
			if queryFree(pos.FromVector(at).Translate(d)) {
				recurse(pos)
			}
		}
	}
	// Find a random starting point.
	startAt := logic.NewVectorFromValues(0, 0, 0)
	for {
		startAt.X, startAt.Y, startAt.Z = rand.Intn(bounds.X), rand.Intn(bounds.Y), rand.Intn(bounds.Z)
		if queryFree(startAt) {
			break
		}
	}
	// Signal the managing routin that this recursor is ready.
	signalNil()
	// Wait for the synch mechanism to tell this to start, then start recursing.
	start.Wait()
	recurse(startAt)
}
