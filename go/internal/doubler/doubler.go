package doubler

import (
	"exercise/internal/state"
	v1 "exercise/pkg/ably/v1"
	"math/big"
)

const (
	// Multiplier to be used for incrementing the value
	Multiplier = 2
)

// GetSequence generates a sequence of values and pushes them to the supplied channel.
func GetSequence(qty int64, seed *big.Int) ([]*big.Int, error) {
	seq := make([]*big.Int, qty)
	initVal := new(big.Int)
	initVal.Set(seed)
	seq[0] = seed

	for cntr := int64(1); cntr < qty; cntr++ {
		initVal.Mul(initVal, big.NewInt(Multiplier))
		value := new(big.Int)
		value.Set(initVal)
		seq[cntr] = value
	}

	return seq, nil
}

// GetRequest build a suitable v1.DoublerRequest from the provided state.Stateful
func GetRequest(s state.Stateful) *v1.Request {
	return &v1.Request{
		Qty:  s.Require() + 1, // +1 to include seed
		Seed: s.Last().Int64(),
	}
}
