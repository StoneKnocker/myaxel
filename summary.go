package main

import (
	"fmt"
	"time"
)

type result struct {
	downLen int64

	finished bool
	start    time.Time
	total    int64
}

func newResult(total int64) *result {
	return &result{
		start: time.Now(),
		total: total,
	}
}

func (r *result) String() string {
	spent := time.Since(r.start)
	desc := "finished"
	if !r.finished {
		desc = "interupted"
	}
	return fmt.Sprintf("%s, fileSize: %d, download %d in %v, %.2f bytes/s", desc, r.total, r.downLen,
		spent, float64(r.downLen)/spent.Seconds())
}
