package lbcluster

import "time"

type Retry struct {
	signal            chan bool
	done              chan bool
	currentCount      int
	maxCount          int
	retryStarted      bool
	maxDuration       time.Duration
	retryDuration     time.Duration
	prevRetryDuration time.Duration
}

const defaultMaxDuration = 5 * time.Minute

func NewRetryModule(retryStartDuration time.Duration) *Retry {
	retry := &Retry{
		maxCount:          -1,
		maxDuration:       defaultMaxDuration,
		retryDuration:     retryStartDuration,
		prevRetryDuration: retryStartDuration,
	}
	return retry
}

func (r *Retry) SetMaxDuration(maxDuration time.Duration) {
	if r.retryStarted {
		return
	}
	r.maxDuration = maxDuration

}

func (r *Retry) SetMaxCount(maxCount int) {
	if r.retryStarted {
		return
	}
	r.maxCount = maxCount
}

func (r *Retry) Start() <-chan bool {
	r.retryStarted = true
	signal := make(chan bool)
	done := make(chan bool)
	end := time.Tick(r.maxDuration)
	r.signal = signal
	r.done = done
	go func() {
		defer close(done)
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
	return r.signal
}

func (r *Retry) run() {
	go func() {
		defer close(r.signal)
		for {
			select {
			case <-r.done:
				return
			default:
				r.signal <- true
				r.currentCount += 1
				time.Sleep(r.retryDuration)
				r.computeNextRetryTime()
			}
		}
	}()
}

// using fibonacci algorithm to compute the next run time
func (r *Retry) computeNextRetryTime() {
	nextDuration := r.retryDuration + r.prevRetryDuration
	r.prevRetryDuration = r.retryDuration
	r.retryDuration = nextDuration
}
