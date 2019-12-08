package utils

import "sync"

type WaitWrapper struct {
	sync.WaitGroup
}

func (w *WaitWrapper) Wrap(fn func()) {
	w.Add(1)
	go func() {
		fn()
		w.Done()
	}()
}
