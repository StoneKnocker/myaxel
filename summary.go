package main

import (
	"fmt"
	"time"
)

type result struct {
	downLen int64
	start   time.Time
	total   int64
}

func newResult(total int64) *result {
	return &result{
		start: time.Now(),
		total: total,
	}
}

func (r *result) finished() bool {
	return r.downLen == r.total
}

func (r *result) String() string {
	desc := "finished"
	if r.downLen != r.total {
		desc = "interrupted"
	}
	return fmt.Sprintf("download %s, target file size: %d bytes, download %d bytes in %s", desc, r.total, r.downLen, time.Since(r.start))
}
