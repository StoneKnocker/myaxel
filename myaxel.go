package main

//TODO signal
//TODO downloader
//todo timeout

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
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

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ok, err := serverSupport(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !ok {
		fmt.Println("server not support for multiple request, you may download it with a web browser")
		return
	}
	fmt.Println("initialing ", outFile, "...")

	f, err := os.Create(outFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	go func() {
		err := <-errChan
		panic(fmt.Sprintf("error collectd: %v", err))
	}()

	summary = newResult(f)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bar := newBar(fileSize, summary)
		bar.show()
	}()

	loader := newDownloader(ctx, fileSize, rawURL)
	loader.do()

	summary.finished = true
	doneChan <- struct{}{}
	wg.Wait()
	fmt.Println(summary)
}

func serverSupport(req *http.Request) (bool, error) {
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
