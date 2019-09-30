package nervo

import (
	"errors"
	"time"
)

//ErrTimeoutReached describes a situation where a certain action took longer than we expected and was aborted because of that
var ErrTimeoutReached error = errors.New("Timeout was reached")

//withTimeout returns an error if the timeout was reached
func withTimeOut(timeout time.Duration, f func()) error {
	t := time.NewTimer(timeout)
	fDoneChan := make(chan struct{})
	go func() {
		f()
		fDoneChan <- struct{}{}
	}()

	select {
	case <-fDoneChan:
		t.Stop()

		return nil
	case <-t.C:
		return ErrTimeoutReached
	}
}
