package lbcluster

import (
	"fmt"
	"time"
)

type Retry struct {
	signal            chan int
	done              chan bool
	tick              chan bool
	currentCount      int
	maxCount          int
	retryStarted      bool
	maxDuration       time.Duration
	retryDuration     time.Duration
	prevRetryDuration time.Duration
	logger            *Log
}

const defaultMaxDuration = 5 * time.Minute

func NewRetryModule(retryStartDuration time.Duration, logger *Log) *Retry {
	retry := &Retry{
		maxCount:          -1,
		currentCount:      0,
		maxDuration:       defaultMaxDuration,
		retryDuration:     retryStartDuration,
		prevRetryDuration: retryStartDuration,
		logger:            logger,
	}
	return retry
}

func (r *Retry) SetMaxDuration(maxDuration time.Duration) error {
	if r.retryStarted {
		return nil
	}
	if maxDuration <= 0 {
		return fmt.Errorf("duration has to be greater than 0")
	}
	r.maxDuration = maxDuration
	return nil
}

func (r *Retry) SetMaxCount(maxCount int) error {
	if r.retryStarted {
		return nil
	}
	if maxCount <= 0 {
		return fmt.Errorf("max count has to be greater than 0")
	}
	r.maxCount = maxCount
	return nil
}

func (r *Retry) start() {
	r.retryStarted = true
	signal := make(chan int)
	done := make(chan bool)
	end := time.Tick(r.maxDuration)
	r.signal = signal
	r.done = done
	go func() {
		for {
			select {
			case <-end:
				r.done <- true
				return
			default:
				if r.currentCount == r.maxCount {
					r.done <- true
					return
				}
			}
		}
	}()
	r.run()
}

func (r *Retry) run() {
	start := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer close(r.done)
		defer close(r.signal)
		defer ticker.Stop()

		for {
			select {
			case <-r.done:
				return
			case <-ticker.C:
				r.nextTick(ticker)
			case <-start:
				r.nextTick(ticker)
			}
		}
	}()
	start <- true
}

func (r *Retry) nextTick(ticker *time.Ticker) {
	r.signal <- r.currentCount + 1
	r.currentCount += 1
	ticker.Reset(r.retryDuration)
	r.computeNextRetryTime()
}

func (r *Retry) Execute(executor func() error) error {
	var err error
	r.start()
	for retryCount := range r.signal {
		err = executor()
		if err != nil {
			r.logger.Debug(fmt.Sprintf("retry count: %v", retryCount))
		} else {
			r.done <- true
		}
	}
	return err
}

// using fibonacci algorithm to compute the next run time
func (r *Retry) computeNextRetryTime() {
	nextDuration := r.retryDuration + r.prevRetryDuration
	r.prevRetryDuration = r.retryDuration
	r.retryDuration = nextDuration
}
