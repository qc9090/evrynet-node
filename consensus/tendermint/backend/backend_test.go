package backend

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evrynet-official/evrynet-client/common"
	"github.com/evrynet-official/evrynet-client/consensus"
	"github.com/evrynet-official/evrynet-client/consensus/tendermint"
	"github.com/evrynet-official/evrynet-client/consensus/tendermint/tests_utils"
	"github.com/evrynet-official/evrynet-client/consensus/tendermint/validator"
	evrynetCore "github.com/evrynet-official/evrynet-client/core"
	"github.com/evrynet-official/evrynet-client/core/types"
	"github.com/evrynet-official/evrynet-client/crypto"
	"github.com/evrynet-official/evrynet-client/ethdb"
	"github.com/evrynet-official/evrynet-client/event"
	"github.com/evrynet-official/evrynet-client/params"
)

func TestSign(t *testing.T) {
	privateKey, _ := tests_utils.GeneratePrivateKey()
	b := &Backend{
		privateKey: privateKey,
	}
	data := []byte("Here is a string....")
	sig, err := b.Sign(data)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
	// Check signature recover
	hashData := crypto.Keccak256([]byte(data))
	pubkey, _ := crypto.Ecrecover(hashData, sig)

	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	if signer != tests_utils.GetAddress() {
		t.Errorf("address mismatch: have %v, want %s", signer.Hex(), tests_utils.GetAddress().Hex())
	}
}

func TestValidators(t *testing.T) {
	var (
		nodePrivateKey = tests_utils.MakeNodeKey()
		nodeAddr       = crypto.PubkeyToAddress(nodePrivateKey.PublicKey)
		validators     = []common.Address{
			nodeAddr,
		}
		genesisHeader = tests_utils.MakeGenesisHeader(validators)
		be            = mustCreateAndStartNewBackend(t, nodePrivateKey, genesisHeader)
	)

	valSet0 := be.Validators(big.NewInt(0))
	assert.Equal(t, 1, valSet0.Size())

	list := valSet0.List()
	log.Println("validator set of block 0 is")

	for _, val := range list {
		log.Println(val.String())
	}

	valSet1 := be.Validators(big.NewInt(1))
	assert.Equal(t, 1, valSet1.Size())

	list = valSet1.List()
	log.Println("validator set of block 1 is")

	for _, val := range list {
		log.Println(val.String())
	}

	valSet2 := be.Validators(big.NewInt(2))
	assert.Equal(t, 0, valSet2.Size())
}

func mustCreateAndStartNewBackend(t *testing.T, nodePrivateKey *ecdsa.PrivateKey, genesisHeader *types.Header) *Backend {
	var (
		address = crypto.PubkeyToAddress(nodePrivateKey.PublicKey)
		trigger = false
		statedb = tests_utils.MustCreateStateDB(t)

		testTxPoolConfig evrynetCore.TxPoolConfig
		blockchain       = &tests_utils.MockChainReader{
			GenesisHeader: genesisHeader,
			MockBlockChain: &tests_utils.MockBlockChain{
				Statedb:       statedb,
				GasLimit:      1000000000,
				ChainHeadFeed: new(event.Feed),
			},
			Address: address,
			Trigger: &trigger,
		}
		pool   = evrynetCore.NewTxPool(testTxPoolConfig, params.TendermintTestChainConfig, blockchain)
		memDB  = ethdb.NewMemDatabase()
		config = tendermint.DefaultConfig
		be     = New(config, nodePrivateKey, WithDB(memDB)).(*Backend)
	)
	statedb.SetBalance(address, new(big.Int).SetUint64(params.Ether))
	defer pool.Stop()
	be.chain = blockchain
	be.currentBlock = blockchain.CurrentBlock

	return be
}

type mockBroadcaster struct {
	handleFn     func(interface{})
	isDisconnect bool
}

// FindPeers returns a map of mockPeer but only one with trigger HandleMsg
func (m *mockBroadcaster) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	if m.isDisconnect {
		return nil
	}
	out := make(map[common.Address]consensus.Peer)
	hasHandle := false
	for addr := range targets {
		if !hasHandle {
			out[addr] = &tests_utils.MockPeer{SendFn: m.handleFn}
			hasHandle = true
			continue
		}
		out[addr] = &tests_utils.MockPeer{}
	}
	return out
}

func (m *mockBroadcaster) Enqueue(id string, block *types.Block) {
	panic("implement me")
}

func TestGossip(t *testing.T) {
	var (
		nodePrivateKey = tests_utils.MakeNodeKey()
		nodeAddr       = crypto.PubkeyToAddress(nodePrivateKey.PublicKey)
		validators     = []common.Address{
			nodeAddr,
		}
		genesisHeader = tests_utils.MakeGenesisHeader(validators)
		be            = mustCreateAndStartNewBackend(t, nodePrivateKey, genesisHeader)

		nodeAddrs = []common.Address{
			common.HexToAddress("1"),
			common.HexToAddress("2"),
			common.HexToAddress("3"),
			nodeAddr,
		}
	)

	dataCh := make(chan string)

	broadcaster := &mockBroadcaster{
		handleFn: func(data interface{}) {
			dataCh <- string(data.([]byte))
		},
		isDisconnect: false,
	}
	be.SetBroadcaster(broadcaster)
	valSet := validator.NewSet(nodeAddrs, tendermint.RoundRobin, 100)

	t.Run("test basic", func(t *testing.T) {
		var expectedData = "aaa"
		err := be.Gossip(valSet, []byte(expectedData))
		require.NoError(t, err)

		select {
		case <-time.After(time.Millisecond * 20):
			t.Fatal("not receive msg to peer")
		case data := <-dataCh:
			assert.Equal(t, expectedData, data)
		}
	})

	t.Run("test retrying broadcast data ", func(t *testing.T) {
		broadcaster.isDisconnect = true
		var expectedData = "aaa"
		err := be.Gossip(valSet, []byte(expectedData))
		require.NoError(t, err)
		select {
		case <-time.After(time.Millisecond * 80):
		case <-dataCh:
			t.Fatal("expected not send to peer when disconnect")
		}

		broadcaster.isDisconnect = false
		select {
		case <-time.After(time.Millisecond * 40):
			t.Fatal("not receive msg to peer")
		case data := <-dataCh:
			assert.Equal(t, expectedData, data)
		}
	})

	t.Run("test skipping retry when having msg", func(t *testing.T) {
		broadcaster.isDisconnect = true
		var expectedData = "aaa"
		err := be.Gossip(valSet, []byte(expectedData))
		require.NoError(t, err)
		select {
		case <-time.After(time.Millisecond * 80):
		case <-dataCh:
			t.Fatal("expected not send to peer when disconnect")
		}

		broadcaster.isDisconnect = false
		var expectedData2 = "bbb"
		err = be.Gossip(valSet, []byte(expectedData2))
		require.NoError(t, err)

		select {
		case <-time.After(time.Millisecond * 40):
			t.Fatal("not receive msg to peer")
		case data := <-dataCh:
			assert.Equal(t, expectedData2, data)
		}
	})
}
