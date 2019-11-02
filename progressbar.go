package main

import (
	"time"

	"github.com/cheggaaa/pb/v3"
)

type bar struct {
	summary *result
	*pb.ProgressBar
}

func newBar(summary *result) *bar {
	return &bar{
		summary:     summary,
		ProgressBar: pb.Default.Start64(summary.total),
	}
}

func (b *bar) show() {
	b.Start()

	for !b.summary.finished {
		b.SetCurrent(b.summary.downLen)
		time.Sleep(time.Millisecond * 200)
	}

	b.SetCurrent(b.summary.total)
	b.Finish()
}
