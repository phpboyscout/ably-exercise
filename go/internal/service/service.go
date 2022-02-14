package service

import (
	"context"
	"crypto/rand"
	"exercise/internal/doubler"
	"exercise/internal/random"
	"exercise/internal/state"
	"math/big"
	"time"

	"google.golang.org/grpc/metadata"

	v1 "exercise/pkg/ably/v1"
)

type Service struct {
	v1.UnimplementedServiceServer
	state map[string]*state.State
}

// seed returns the seed if greater than zero otherwise returns a random integer between 0 and MaxSeed.
func (s *Service) seed(seed int64) *big.Int {
	if seed > 0 {
		return big.NewInt(seed)
	}

	max := int64(MaxSeed)

	r, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		logger.Error().Err(err)
	}

	return r
}

// getState retrieves/instantiates a state object for a request.
// if a client_id is supplied in metadata then the state is persisted in memory
func (s *Service) getState(ctx context.Context, qty int64, seq []*big.Int) *state.State {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if clientID, ok := md["client-id"]; ok {
			if state, ok := s.state[clientID[0]]; ok {
				return state
			}
			s.state[clientID[0]] = state.NewState(qty, seq)

			return s.state[clientID[0]]
		}
	}

	return state.NewState(qty, seq)
}

// Doubler handles the incoming request and pushes values into the return stream.
func (s *Service) Doubler(req *v1.Request, stream v1.Service_DoublerServer) error {
	seq, _ := doubler.GetSequence(req.GetQty(), big.NewInt(req.GetSeed()))
	state := s.getState(stream.Context(), req.GetQty(), seq)

	for {
		if !state.Next() {
			break
		}
		err := stream.Send(&v1.Response{Value: state.Current().Bytes()})
		if err != nil {
			logger.Error().Err(err)
		}
		time.Sleep(interval)
	}
	err := stream.Send(&v1.Response{Checksum: state.Total().Bytes()})
	if err != nil {
		logger.Error().Err(err)
	}
	stream.Context().Done()

	return nil
}

func (s *Service) Random(req *v1.Request, stream v1.Service_RandomServer) error {
	seq, err := random.GetSequence(req.GetQty(), big.NewInt(0))
	if err != nil {
		return err
	}
	state := s.getState(stream.Context(), req.GetQty(), seq)

	for {
		err := stream.Send(&v1.Response{Value: state.Current().Bytes()})
		if err != nil {
			logger.Error().Err(err)
		}
		if !state.Next() {
			break
		}
		time.Sleep(interval)
	}
	err = stream.Send(&v1.Response{Checksum: state.Total().Bytes()})
	if err != nil {
		logger.Error().Err(err)
	}
	stream.Context().Done()

	return nil
}

// MaintainStates provides a convenience method to clean up stale states.
func (s *Service) MaintainStates() {
	for {
		time.Sleep(time.Duration(1) * time.Second)
		now := time.Now()
		for id, st := range s.state {
			if now.Sub(st.Accessed()).Seconds() > StateTTL {
				delete(s.state, id)
			} else {
				logger.Debug().
					Str("client-id", id).
					Int64("position", st.Position()).
					Float64("ttl", StateTTL-now.Sub(st.Accessed()).Seconds()).
					Send()
			}
		}
	}
}

// NewService instantiates a new service container.
func NewService() *Service {
	return &Service{
		state: map[string]*state.State{},
	}
}
