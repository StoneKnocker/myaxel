package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
)

//Downloader struct
type Downloader struct {
	sync.Mutex
	fl *os.File

	routineNum int
	url        string
	ctx        context.Context
	summary    *Summary
}

//NewDownloader for file download
func NewDownloader(ctx context.Context, routineNum int, filename string, url string, summary *Summary) *Downloader {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	return &Downloader{
		fl:         f,
		routineNum: routineNum,
		url:        url,
		ctx:        ctx,
		summary:    summary,
	}
}

func (d *Downloader) makeRequest(routineNO int) (*http.Request, error) {
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
	req.Header.Set("Range", rangeStr)

	return req, nil
}

//Do starts to download file
func (d *Downloader) Do() {
	var wg sync.WaitGroup
	wg.Add(d.routineNum)
	for i := 0; i < d.routineNum; i++ {
		go func(i int) {
			defer wg.Done()

			req, err := d.makeRequest(i)
			if err != nil {
				errCollecter(err)
				return
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				errCollecter(err)
				return
			}
			defer resp.Body.Close()

			rangeStart := int64(i) * d.summary.total / int64(d.routineNum)

			var seekLen int64
			for {
				select {
				case <-d.ctx.Done():
					errCollecter(d.ctx.Err())
					return
				default:
					d.Lock()
					d.fl.Seek(rangeStart+seekLen, os.SEEK_SET)
					written, err := io.CopyN(d.fl, resp.Body, 4096)
					if err != nil {
						if err != io.EOF {
							d.Unlock()
							errCollecter(err)
							return
						}
						d.Unlock()
						atomic.AddInt64(&d.summary.downLen, written)
						return
					}
					d.Unlock()
					atomic.AddInt64(&d.summary.downLen, written)

					seekLen += written
				}
			}
		}(i)
	}
	wg.Wait()
	d.fl.Close()
}
