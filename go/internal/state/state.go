package state

import (
	"math/big"
	"sync/atomic"
	"time"
)

type Stateful interface {
	Position() int64
	Quantity() int64
	Require() int64
	Current() *big.Int
	Last() *big.Int
	Next() bool
	Set([]*big.Int)
	Add(...*big.Int)
	Total() *big.Int
	Accessed() time.Time
	Sequence() []*big.Int
}

// State contains the given state of a grpc request containing the last value generated and the number of values
// seen by the state as well as materialising the number of values generated and the sum of those values.
type State struct {
	// quantity of values requested.
	qty int64
	// a cursor for tracking the current value when iterating.
	cursor int64
	// last value to have been processed.
	sequence []*big.Int
	// accessed time the state was last accessed.
	accessed time.Time
}

// Position returns the value of the cursor
func (s *State) Position() int64 {
	return s.cursor
}

// Seek sets the cursort to the specified position
func (s *State) Seek(position int64) {
	s.accessed = time.Now()
	s.cursor = position
}

// Quantity returns the originally requested quantity for this state.
func (s *State) Quantity() int64 {
	return s.qty
}

// Require advises how many values are required to meet requested quantity.
func (s *State) Require() int64 {
	return s.qty - int64(len(s.sequence))
}

// Current returns the latest value to have been added to state.
func (s *State) Current() *big.Int {
	s.accessed = time.Now()
	return s.sequence[s.cursor]
}

// Set the sequence, also resets the cursor.
func (s *State) Set(seq []*big.Int) {
	s.sequence = seq
	s.cursor = 0
	s.accessed = time.Now()
}

// Add a new number to the sequence.
func (s *State) Add(values ...*big.Int) {
	s.sequence = append(s.sequence, values...)
	s.accessed = time.Now()
}

// Last returns the last item in the sequence, this does not move the cursor.
func (s *State) Last() *big.Int {
	s.accessed = time.Now()
	return s.sequence[len(s.sequence)-1]
}

// Next moves the cursor forward in the sequence by one and
// returns true if the cursor remains within range of the sequence
// returns false if out of range of the sequence.
func (s *State) Next() bool {
	atomic.AddInt64(&s.cursor, 1)
	s.accessed = time.Now()
	return s.cursor < int64(len(s.sequence))
}

// Total generates the sum total of all values in the sequence.
func (s *State) Total() *big.Int {
	total := big.NewInt(0)
	for _, s := range s.sequence {
		total.Add(total, s)
	}
	s.accessed = time.Now()
	return total
}

// Sequence returns the entire sequence currently stored.
func (s *State) Sequence() []*big.Int {
	s.accessed = time.Now()
	return s.sequence
}

// Accessed returns the timme the state was last accessed.
func (s *State) Accessed() time.Time {
	return s.accessed
}

// NewState instantiate a new state object
func NewState(qty int64, seq []*big.Int) *State {
	return &State{
		qty:      qty,
		cursor:   int64(0),
		sequence: seq,
		accessed: time.Now(),
	}
}
