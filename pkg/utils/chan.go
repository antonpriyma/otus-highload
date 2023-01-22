package utils

import "time"

type ChanLocker interface {
	Lock() bool
	Unlock()
}

func NewChanLocker() ChanLocker {
	return chanLocker{
		ch: make(chan struct{}, 1),
	}
}

type chanLocker struct {
	ch chan struct{}
}

func (c chanLocker) Lock() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (c chanLocker) Unlock() {
	select {
	case <-c.ch:
	default:
		panic("unlock on unlocked locker")
	}
}

func IsSignalChanClosed(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func SleepOrChannelClose(ch chan struct{}, t time.Duration) {
	if t == 0 {
		return
	}

	sleepTimer := time.NewTimer(t)
	defer sleepTimer.Stop()

	select {
	case <-ch:
		return
	case <-sleepTimer.C:
		return
	}
}
