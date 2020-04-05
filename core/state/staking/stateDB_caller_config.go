package staking

// Constants represents the configuration name of all state variables.
const (
	WithdrawsStateIndexName    = "withdrawsState"
	CandidateVotersIndexName   = "candidateVoters"
	CandidateDataIndexName     = "candidateData"
	CandidatesIndexName        = "candidates"
	StartBlockIndexName        = "startBlock"
	EpochPeriodIndexName       = "epochPeriod"
	MaxValidatorSizeIndexName  = "maxValidatorSize"
	MinValidatorStakeIndexName = "minValidatorStake"
	MinVoterCapIndexName       = "minVoterCap"
	AdminIndexName             = "admin"
)

// StorageLayout represents the struct of object its get from a json data file
type StorageLayout struct {
	Label  string `json:"label"`
	Offset uint16 `json:"offset"`
	Slot   uint64 `json:"slot,string"`
}

// LayOut represents the Offset and Slot order of a state variable
type LayOut struct {
	Offset uint16
	Slot   uint64
}

// IndexConfigs represents the configuration index of state variables.
type IndexConfigs struct {
	WithdrawsStateLayout    LayOut //1
	CandidateVotersLayout   LayOut //2
	CandidateDataLayout     LayOut //3
	CandidatesLayout        LayOut //4
	StartBlockLayout        LayOut //5
	EpochPeriodLayout       LayOut //6
	MaxValidatorSizeLayout  LayOut //7
	MinValidatorStakeLayout LayOut //8
	MinVoterCapLayout       LayOut //9
	AdminLayout             LayOut //10
}

// DefaultConfig represents he default configuration.
var DefaultConfig = &IndexConfigs{
	WithdrawsStateLayout:    NewLayOut(1, 0),
	CandidateVotersLayout:   NewLayOut(2, 0),
	CandidateDataLayout:     NewLayOut(3, 0),
	CandidatesLayout:        NewLayOut(4, 0),
	StartBlockLayout:        NewLayOut(5, 0),
	EpochPeriodLayout:       NewLayOut(6, 0),
	MaxValidatorSizeLayout:  NewLayOut(7, 0),
	MinValidatorStakeLayout: NewLayOut(8, 0),
	MinVoterCapLayout:       NewLayOut(9, 0),
	AdminLayout:             NewLayOut(10, 0),
}

// NewLayOut returns new instance of a LayOut
func NewLayOut(slot uint64, offset uint16) LayOut {
	return LayOut{Offset: offset, Slot: slot}
}
