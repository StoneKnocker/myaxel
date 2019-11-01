package main

import (
	"github.com/cheggaaa/pb/v3"
)

type bar struct {
	summary *result
	*pb.ProgressBar
}

func newBar(count int64, summary *result) *bar {
	return &bar{
		summary:     summary,
		ProgressBar: pb.Default.Start64(count),
	}
}

func (b *bar) show() {
	b.Start()

loop:
	for {
		select {
		case <-doneChan:
			break loop
		default:
			b.SetCurrent(summary.downLen)
		}
	}

	b.Finish()
}
