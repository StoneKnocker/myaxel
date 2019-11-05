package main

import (
	"fmt"
	"time"
)

//Summary struct
type Summary struct {
	downLen int64
	start   time.Time
	total   int64
}

//NewSummary return new summary
func NewSummary(total int64) *Summary {
	return &Summary{
		start: time.Now(),
		total: total,
	}
}

//Finished reports whether file download finished
func (r *Summary) Finished() bool {
	return r.downLen == r.total
}

//String returns the summary info
func (r *Summary) String() string {
	desc := "finished"
	if r.downLen != r.total {
		desc = "interrupted"
	}
	return fmt.Sprintf("download %s, target file size: %d bytes, download %d bytes in %s", desc, r.total, r.downLen, time.Since(r.start))
}
