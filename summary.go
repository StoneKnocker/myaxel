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
	desc := "finished"
	if !r.finished {
		desc = "interrupted"
	}
	return fmt.Sprintf("download %s, fileSize: %d bytes, download %d bytes in %s", desc, r.total, r.downLen, time.Since(r.start))
}
