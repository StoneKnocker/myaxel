package main

import (
	"time"

	"github.com/cheggaaa/pb/v3"
)

//Bar struct
type Bar struct {
	summary *Summary
	*pb.ProgressBar
}

//NewBar for bar
func NewBar(summary *Summary) *Bar {
	b := &Bar{
		summary:     summary,
		ProgressBar: pb.New64(summary.total),
	}
	b.Set(pb.Bytes, true)
	return b
}

//Show the progress bar
func (b *Bar) Show() {
	b.Start()

	for !b.summary.Finished() {
		b.SetCurrent(b.summary.downLen)
		time.Sleep(time.Millisecond)
	}

	b.SetCurrent(b.summary.total)
	b.Finish()
}
