package main

//TODO signal
//TODO downloader
//todo timeout
//todo fix ./myaxel https://www.baidu.com progress bar

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	timeout  time.Duration
	outFile  string
	insecure bool
	fileSize int64
	summary  *result

	httpClient = http.DefaultClient

	errChan  = make(chan error)
	doneChan = make(chan struct{})
)

func init() {
	flag.StringVar(&outFile, "o", "default", "local output file name")
	flag.DurationVar(&timeout, "T", 30*time.Minute, "timeout")
	flag.BoolVar(&insecure, "k", false, "do not verify the SSL certificate")
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	if insecure {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	go func() {
		err := <-errChan
		panic(err)
	}()

	rawURL := flag.Arg(0)
	filename, err := parseFilename(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	if outFile != "" {
		outFile = filename
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		showError(err)
	}

	ok, err := multiSupport(req)
	if err != nil {
		showError(err)
	}
	if !ok {
		fmt.Println("server not support or multiple thread, you may download it with a web browser")
		return
	}
	fmt.Println("initialing ", outFile, "...")

	f, err := os.Create(outFile)
	if err != nil {
		showError(err)
	}
	summary = newResult(f)
	defer summary.f.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bar := newBar(fileSize, summary)
		bar.show()
	}()
	multiDownlad(req, summary.f)

	summary.finished = true
	doneChan <- struct{}{}
	wg.Wait()
	fmt.Println(summary)
}

func multiDownlad(req *http.Request, f *os.File) {
	numGoroutine := runtime.NumCPU()
	rangeSize := fileSize / int64(numGoroutine)

	var wg sync.WaitGroup
	wg.Add(numGoroutine)
	for i := 0; i < numGoroutine; i++ {
		go func(i int) {
			defer wg.Done()
			rangeStart := int64(i) * rangeSize
			rangeEnd := rangeSize*int64(i+1) - 1
			if i == numGoroutine-1 {
				rangeEnd = fileSize
			}
			rangeStr := fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd)
			newReq := cloneRequest(req)
			newReq.Header.Set("Range", rangeStr)

			resp, err := httpClient.Do(newReq)
			if err != nil {
				showError(err)
				return
			}
			defer resp.Body.Close()

			var seekLen int64
			for {
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
		}(i)
	}
	wg.Wait()
}

func multiSupport(req *http.Request) (bool, error) {
	resp, err := http.Head(req.URL.String())
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New("can't fetch the file")
	}
	if resp.Header.Get("Content-Length") == "" {
		return false, errors.New("can't get the file length")
	}
	fileSize = resp.ContentLength

	newReq := cloneRequest(req)
	newReq.Header.Set("Range", "Bytes=0-1")
	resp, err = httpClient.Do(newReq)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 206 {
		return false, nil
	}
	return true, nil
}

func cloneRequest(req *http.Request) *http.Request {
	newReq := new(http.Request)
	*newReq = *req
	newReq.Header = make(http.Header, len(req.Header))
	for k, v := range req.Header {
		newReq.Header[k] = append([]string(nil), v...)
	}

	return newReq
}

func showError(err error) {
	panic(err)
}
