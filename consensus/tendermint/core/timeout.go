package core

import (
	"math/big"
	"time"

	"github.com/evrynet-official/evrynet-client/log"
)

const (
	tickTockBufferSize = 10
)

// TimeoutTicker is a timer that schedules timeouts
// conditional on the height/round/step in the timeoutInfo.
// The timeoutInfo.Duration may be non-positive.
type TimeoutTicker interface {
	Start() error
	Stop() error
	Chan() <-chan timeoutInfo       // on which to receive a timeout
	ScheduleTimeout(ti timeoutInfo) // reset the timer
}

// timeoutInfo keep track about a timeout job
type timeoutInfo struct {
	Duration    time.Duration `json:"duration"`
	BlockNumber *big.Int      `json:"block_number"`
	Round       int64         `json:"round"`
	Step        RoundStepType `json:"step"`
}

// earlierOrEqual return true if timeoutInfo A is earlier or equal than timeoutInfo B
// otherwise it return false
// a timeoutInfo A is said to be earlier Or Equal than timeoutInfo B if:
// A.BlockNumber < B.BlockNumber || (A.BlockNumber == B.BlockNumber && A.Round< B.Round)  || (A.BlockNumber == B.BlockNumber && A.Round == B.Round && A.Step<= B.Step)
func (A timeoutInfo) earlierOrEqual(B timeoutInfo) bool {
	if A.BlockNumber.Cmp(B.BlockNumber) < 0 {
		return true
	}

	if A.BlockNumber.Cmp(B.BlockNumber) == 0 {
		if A.Round < B.Round {
			return true
		}
		if A.Round == B.Round && (A.Step > 0 && A.Step <= B.Step) {
			return true
		}
	}

	return false
}

// timeoutTicker wraps time.Timer, and implements TimeoutTicker
// scheduling timeouts only for greater height/round/step
// than what it's already seen.
// Timeouts are scheduled along the tickChan,
// and fired on the tockChan.
// NOTE: timeoutTicker only allow 1 timeout to run at a time, any newer timeout will stop the earlier one.
type timeoutTicker struct {
	started  bool
	timer    *time.Timer
	tickChan chan timeoutInfo // for scheduling timeouts
	tockChan chan timeoutInfo // for notifying about them
	Quit     <-chan struct{}
}

// NewTimeoutTicker returns a new TimeoutTicker that's ready to use
func NewTimeoutTicker() TimeoutTicker {
	//TODO: allow caller to indicate buffer size
	tt := &timeoutTicker{
		timer:    time.NewTimer(time.Duration(1<<63 - 1)),
		tickChan: make(chan timeoutInfo, tickTockBufferSize),
		tockChan: make(chan timeoutInfo, tickTockBufferSize),
	}
	return tt
}

func (tt *timeoutTicker) Start() error {
	go tt.timeoutRoutine()
	return nil
}

func (tt *timeoutTicker) Stop() error {
	tt.stopTimer()
	return nil
}

// ScheduleTimeout schedules a new timeout by sending on the internal tickChan.
// The timeoutRoutine is always available to read from tickChan, so this won't block.
// The scheduling may fail if the timeoutRoutine has already scheduled a timeout for a later height/round/step.
func (tt *timeoutTicker) ScheduleTimeout(ti timeoutInfo) {
	tt.tickChan <- ti
}

// Chan returns a channel on which timeouts are sent.
func (tt *timeoutTicker) Chan() <-chan timeoutInfo {
	return tt.tockChan
}

// stop the timer and drain if necessary
func (tt *timeoutTicker) stopTimer() {
	// Stop() returns false if it was already fired or was stopped
	if !tt.timer.Stop() {
		select {
		case <-tt.timer.C:
		default:
			log.Debug("Timer already stopped")
		}
	}
}

// send on tickChan to start a new timer.
// timers are interupted and replaced by new ticks from later steps
// timeouts of 0 on the tickChan will be immediately relayed to the tockChan
func (tt *timeoutTicker) timeoutRoutine() {
	var ti = timeoutInfo{
		BlockNumber: big.NewInt(0),
		Round:       0,
	}
	//TODO: DO we need mutex for this?
	for {
		select {
		case newti := <-tt.tickChan:

			// ignore tickers for old height/round/step
			if newti.earlierOrEqual(ti) {
				continue
			}
			// stop the last timer
			tt.stopTimer()

			// update timeoutInfo and reset timer
			// NOTE time.Timer allows duration to be non-positive
			ti = newti
			tt.timer.Reset(ti.Duration)
			log.Debug("Scheduled timeout", "dur", ti.Duration, "height", ti.BlockNumber, "round", ti.Round, "step", ti.Step)
		case <-tt.timer.C:
			log.Info("Timed out", "dur", ti.Duration, "height", ti.BlockNumber, "round", ti.Round, "step", ti.Step)
			// go routine here guarantees timeoutRoutine doesn't block.
			// Determinism comes from playback in the handleEvents.
			// We can eliminate it by merging the timeoutRoutine into receiveRoutine
			//  and managing the timeouts ourselves with a millisecond ticker
			// TODO: see if we can fire directly into core.events
			go func(toi timeoutInfo) { tt.tockChan <- toi }(ti)
		case <-tt.Quit:
			return
		}
	}
}