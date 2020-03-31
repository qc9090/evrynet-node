package staking

import (
	"math/big"
	"sort"
	"strings"

	"github.com/Evrynetlabs/evrynet-node/accounts/abi/bind"
	"github.com/Evrynetlabs/evrynet-node/common"
	"github.com/Evrynetlabs/evrynet-node/core/state"
	"github.com/Evrynetlabs/evrynet-node/crypto"
)

// stateDBStakingCaller creates a wrapper with statedb to implements ContractCaller
type stateDBStakingCaller struct {
	stateDB *state.StateDB
	config  *Config
}

// NewStateDbStakingCaller return instance of StakingCaller which reads data directly from state DB
func NewStateDbStakingCaller(state *state.StateDB, cfg *Config) StakingCaller {
	return &stateDBStakingCaller{
		stateDB: state,
		config:  cfg,
	}
}

// GetCandidateStake returns current stake of a candidate
func (c *stateDBStakingCaller) GetCandidateStake(scAddress common.Address, candidate common.Address) *big.Int {
	locValidatorsState := getLocMappingAtKey(candidate.Hash(), c.config.CandidateDataIndex)
	//TODO: change uint64(1) into a constant
	locStake := locValidatorsState.Add(locValidatorsState, new(big.Int).SetUint64(uint64(1)))
	stake := c.stateDB.GetState(scAddress, common.BigToHash(locStake))
	return stake.Big()
}

// GetValidators returns validators from stateDB and block number of the caller by smart-contract's address
func (c *stateDBStakingCaller) GetValidators(scAddress common.Address) ([]common.Address, error) {
	// check if this address is a valid contract, this will help us return better error
	codes := c.stateDB.GetCode(scAddress)
	if len(codes) == 0 {
		return nil, bind.ErrNoCode
	}

	candidates, err := c.getCandidates(scAddress)
	if err != nil {
		return nil, err
	}

	var (
		validators []common.Address
		stakes     = make(map[common.Address]*big.Int)
	)
	minValStake := c.getMinValidatorStake(scAddress)
	for _, candidate := range candidates {
		stake := c.GetCandidateStake(scAddress, candidate)
		if stake.Cmp(minValStake) < 0 {
			continue
		}

		validators = append(validators, candidate)
		stakes[candidate] = stake
	}

	maxValSize := c.getMaxValidatorSize(scAddress)
	//TODO: reuse this block of code with evmStakingCaller
	if len(validators) <= maxValSize {
		return validators, nil
	}

	sort.Slice(validators, func(i, j int) bool {
		if stakes[validators[i]].Cmp(stakes[validators[j]]) == 0 {
			return strings.Compare(validators[i].String(), validators[j].String()) > 0
		}
		return stakes[validators[i]].Cmp(stakes[validators[j]]) > 0
	})

	return candidates[:maxValSize], nil
}

func (c *stateDBStakingCaller) GetValidatorsData(common.Address, []common.Address) (map[common.Address]CandidateData, error) {
	panic("implement me")
}

func (c *stateDBStakingCaller) getMaxValidatorSize(scAddress common.Address) int {
	slotHash := common.BigToHash(new(big.Int).SetUint64(c.config.MaxValidatorSizeIndex))
	ret := c.stateDB.GetState(scAddress, slotHash)
	return int(ret.Big().Int64())
}

func (c *stateDBStakingCaller) getMinValidatorStake(scAddress common.Address) *big.Int {
	slotHash := common.BigToHash(new(big.Int).SetUint64(c.config.MinValidatorStakeIndex))
	ret := c.stateDB.GetState(scAddress, slotHash)
	return ret.Big()
}

func (c *stateDBStakingCaller) getCandidates(scAddress common.Address) ([]common.Address, error) {
	slot := c.config.CandidatesIndex
	slotHash := common.BigToHash(new(big.Int).SetUint64(slot))
	arrLength := c.stateDB.GetState(scAddress, slotHash)
	if arrLength.Big().Cmp(big.NewInt(0)) == 0 {
		return nil, ErrEmptyValidatorSet
	}
	var candidates []common.Address
	for i := uint64(0); i < arrLength.Big().Uint64(); i++ {
		key := getLocDynamicArrAtElement(slotHash, i, 1)
		ret := c.stateDB.GetState(scAddress, key)
		candidates = append(candidates, common.HexToAddress(ret.Hex()))
	}
	return candidates, nil
}

func getLocDynamicArrAtElement(slotHash common.Hash, index uint64, elementSize uint64) common.Hash {
	slotKecBig := crypto.Keccak256Hash(slotHash.Bytes()).Big()
	arrBig := slotKecBig.Add(slotKecBig, new(big.Int).SetUint64(index*elementSize))
	return common.BigToHash(arrBig)
}

func getLocMappingAtKey(key common.Hash, slot uint64) *big.Int {
	slotHash := common.BigToHash(new(big.Int).SetUint64(slot))
	retByte := crypto.Keccak256(key.Bytes(), slotHash.Bytes())
	ret := new(big.Int)
	ret.SetBytes(retByte)
	return ret
}