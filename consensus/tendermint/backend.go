package tendermint

import (
	"math/big"

	"github.com/evrynet-official/evrynet-client/common"
	"github.com/evrynet-official/evrynet-client/common/hexutil"
	"github.com/evrynet-official/evrynet-client/consensus"
	"github.com/evrynet-official/evrynet-client/core/types"
	"github.com/evrynet-official/evrynet-client/event"
)

var (
	// Magic nonce number to vote on adding a new validator
	NonceAuthVote = hexutil.MustDecode("0xffffffffffffffff")
	// Magic nonce number to vote on removing a validator.
	NonceDropVote = hexutil.MustDecode("0x0000000000000000")
)

// Backend provides application specific functions for Tendermint core
type Backend interface {
	// Address returns the Ethereum address of the node running this backend
	Address() common.Address

	// EventMux returns the event mux used for Core to subscribe/ send events back to Backend.
	// Think of it as pub/sub models
	EventMux() *event.TypeMux

	// Sign signs input data with the backend's private key
	Sign([]byte) ([]byte, error)

	// Gossip sends a message to all validators (exclude self)
	// these message are send via p2p network interface.
	Gossip(valSet ValidatorSet, payload []byte) error

	// Broadcast sends a message to all validators (including self)
	// It will call gossip and post an identical event to its EventMux().
	Broadcast(valSet ValidatorSet, payload []byte) error

	// Validators returns the validator set
	Validators(blockNumber *big.Int) ValidatorSet

	// CurrentHeadBlock get the current block of from the canonical chain.
	CurrentHeadBlock() *types.Block

	// FindExistingPeers check validator peers exist or not by address
	FindExistingPeers(targets ValidatorSet) map[common.Address]consensus.Peer

	//Commit send the consensus block back to miner, it should also handle the logic after a block get enough vote to be the next block in chain
	Commit(block *types.Block)

	//EnqueueBlock adds the block returned from consensus into fetcher queue to update the chain to that specific block.
	EnqueueBlock(block *types.Block)

	// GetValidators returns val-set from snapshot
	GetValidators(blockNumber *big.Int, chain consensus.ChainReader) ValidatorSet
}
