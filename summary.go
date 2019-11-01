package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type result struct {
	sync.Mutex
	f        *os.File
	downLen  int64
	finished bool

	start time.Time
}

func newResult(f *os.File) *result {
	return &result{f: f}
}

func (r *result) String() string {
	timeSpent := time.Since(r.start)

	desc := "finished."
	if !r.finished {
		desc = "interupted!"
	}
	return fmt.Sprintf("%s, fileSize: %d, download %d in %v, %.2f bytes/s", desc, fileSize, r.downLen, timeSpent, float64(fileSize)/timeSpent.Seconds())
}
