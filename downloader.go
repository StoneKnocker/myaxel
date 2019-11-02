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
)

type downloader struct {
	totalSize  int64
	routineNum int
	url        string
	ctx        context.Context
}

func newDownloader(ctx context.Context, fileSize int64, url string) *downloader {
	return &downloader{
		totalSize:  fileSize,
		routineNum: runtime.NumCPU(),
		url:        url,
		ctx:        ctx,
	}
}

func (d *downloader) makeRequest(routineNO int) (*http.Request, error) {
	req, err := http.NewRequestWithContext(d.ctx, "GET", d.url, nil)
	if err != nil {
		return nil, err
	}
	rangeSize := d.totalSize / int64(d.routineNum)
	rangeStart := int64(routineNO) * rangeSize
	rangeEnd := rangeSize*int64(routineNO+1) - 1
	if routineNO == d.routineNum-1 {
		rangeEnd = d.totalSize
	}
	rangeStr := fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd)
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

			rangeStart := int64(i) * d.totalSize / int64(d.routineNum)

			var seekLen int64
			for {
				select {
				case <-d.ctx.Done():
					errChan <- errors.New("timeout, downloader exit")
					return
				case <-doneChan:
					return
				default:
					summary.Lock()
					summary.f.Seek(rangeStart+seekLen, os.SEEK_SET)
					written, err := io.CopyN(summary.f, resp.Body, 4096)
					if err != nil {
						if err != io.EOF {
							errChan <- err
							summary.Unlock()
							return
						}
						summary.downLen += written
						summary.Unlock()
						break
					}

					seekLen += written
					summary.downLen += written
					summary.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()
}
