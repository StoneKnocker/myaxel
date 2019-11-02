package main

import (
	"fmt"
	"time"

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

	for !summary.finished {
		b.SetCurrent(summary.downLen)
		time.Sleep(time.Millisecond * 200)
	}

	b.SetCurrent(b.total)
	b.Finish()
	fmt.Println("finished bar ")
}
