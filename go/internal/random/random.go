package random

import (
	"crypto/rand"
	"exercise/internal/state"
	v1 "exercise/pkg/ably/v1"
	"math/big"
)

const MaxValue = int64(^uint32(0))

// GetSequence generates a sequence of values and pushes them to the supplied channel.
func GetSequence(qty int64, _ *big.Int) ([]*big.Int, error) {
	seq := make([]*big.Int, qty+1)

	for cntr := int64(0); cntr < qty+1; cntr++ {
		r, err := rand.Int(rand.Reader, big.NewInt(MaxValue))
		if err != nil {
			return nil, err
		}
		seq[cntr] = r
	}

	return seq, nil
}

// GetRequest build a suitable v1.DoublerRequest from the provided state.Stateful
func GetRequest(s state.Stateful) *v1.Request {
	return &v1.Request{
		Qty:  s.Require(),
		Seed: s.Last().Int64(),
	}
}
