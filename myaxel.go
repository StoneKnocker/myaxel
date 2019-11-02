package main

//TODO signal
//TODO downloader
//todo add unit test

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

	errChan = make(chan error)
	sigChan = make(chan os.Signal, 1)
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

	//deal with signal
	signal.Notify(sigChan, os.Interrupt)
	go signalHandler()

	//deal with error
	go func() {
		err := <-errChan
		panic(fmt.Sprintf("error collected: %v", err))
	}()

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

	ok, err := serverSupport(rawURL)
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
	summary = newResult(f)

	go func() {
		loader := newDownloader(ctx, fileSize, rawURL, summary)
		loader.do()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bar := newBar(fileSize, summary)
		bar.show()
	}()
	wg.Wait()

	fmt.Println(summary)
}

func serverSupport(rawURL string) (bool, error) {
	resp, err := http.Head(rawURL)
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

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Range", "Bytes=0-1")
	resp, err = httpClient.Do(req)
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

func signalHandler() {
	switch <-sigChan {
	case os.Interrupt:
		fmt.Println("interrupted!!")
		os.Exit(1)
		panic(summary)
	default:
		panic("unknown")
	}
}
