package main

import (
	"github.com/cheggaaa/pb/v3"
)

type bar struct {
	summary *result
	total   int64
	*pb.ProgressBar
}

func newBar(count int64, summary *result) *bar {
	return &bar{
		summary:     summary,
		total:       count,
		ProgressBar: pb.Default.Start64(count),
	}
}

func (b *bar) show() {
	b.Start()

loop:
	for {
		select {
		case <-doneChan:
			b.SetCurrent(b.total)
			break loop
		default:
			b.SetCurrent(summary.downLen)
		}
	}

	b.Finish()
}
