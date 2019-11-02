package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

type downloader struct {
	sync.Mutex
	fl *os.File

	routineNum int
	url        string
	ctx        context.Context
	summary    *result
}

func newDownloader(ctx context.Context, filename string, url string, summary *result) *downloader {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	return &downloader{
		fl:         f,
		routineNum: runtime.NumCPU(),
		url:        url,
		ctx:        ctx,
		summary:    summary,
	}
}

func (d *downloader) makeRequest(routineNO int) (*http.Request, error) {
	req, err := http.NewRequestWithContext(d.ctx, "GET", d.url, nil)
	if err != nil {
		return nil, err
	}
	rangeSize := d.summary.total / int64(d.routineNum)
	rangeStart := int64(routineNO) * rangeSize
	rangeEnd := rangeSize*int64(routineNO+1) - 1
	if routineNO == d.routineNum-1 {
		rangeEnd = d.summary.total
	}
	rangeStr := fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd)
	fmt.Println(rangeStr)
	req.Header.Set("Range", rangeStr)

	return req, nil
}

func (d *downloader) do() {
	var wg sync.WaitGroup
	wg.Add(d.routineNum)
	for i := 0; i < d.routineNum; i++ {
		go func(i int) {
			defer wg.Done()

			req, err := d.makeRequest(i)
			if err != nil {
				errChan <- err
				return
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()

			rangeStart := int64(i) * d.summary.total / int64(d.routineNum)

			var seekLen int64
			for {
				select {
				case <-d.ctx.Done():
					errChan <- errors.New("timeout, downloader exit")
					return
				default:
					d.Lock()
					d.fl.Seek(rangeStart+seekLen, os.SEEK_SET)
					written, err := io.CopyN(d.fl, resp.Body, 4096)
					if err != nil {
						if err != io.EOF {
							errChan <- err
							d.Unlock()
							return
						}
						d.Unlock()
						atomic.AddInt64(&d.summary.downLen, written)
						strChan <- fmt.Sprintf("thread num: %d, down len: %d", i, seekLen+written)
						return
					}
					d.Unlock()

					seekLen += written
					atomic.AddInt64(&d.summary.downLen, written)
				}
			}
		}(i)
	}
	wg.Wait()
	info, _ := d.fl.Stat()
	_ = info
	d.fl.Close()
	d.summary.finished = true
}
