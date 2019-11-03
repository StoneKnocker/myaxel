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
	b := &bar{
		summary:     summary,
		ProgressBar: pb.New64(summary.total),
	}
	b.Set(pb.Bytes, true)
	return b
}

func (b *bar) show() {
	b.Start()

	for !b.summary.finished {
		b.SetCurrent(b.summary.downLen)
		time.Sleep(time.Millisecond * 50)
	}

	b.SetCurrent(b.summary.total)
	b.Finish()
}
