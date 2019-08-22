package tendermint

import (
	"time"

	"github.com/evrynet-official/evrynet-client/common"
	"github.com/evrynet-official/evrynet-client/event"
)

// Backend provides application specific functions for Tendermint core
type Backend interface {
	// Address returns the Ethereum address of the node running this backend
	Address() common.Address

	// EventMux returns the event mux in backend
	EventMux() *event.TypeMux

	// Sign signs input data with the backend's private key
	Sign([]byte) ([]byte, error)

	// Gossip sends a message to all validators (exclude self)
	Gossip(valSet ValidatorSet, payload []byte) error

	// Broadcast sends a message to all validators (including self)
	Broadcast(valSet ValidatorSet, payload []byte) error

	// Verify verifies the proposal. If a consensus.ErrFutureBlock error is returned,
	// the time difference of the proposal and current time is also returned.
	Verify(Proposal) (time.Duration, error)
}
