package main

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
	"syscall"
	"time"
)

var (
	timeout  time.Duration
	filename string
	insecure bool

	filesize   int64
	summary    *result
	httpClient = http.DefaultClient

	errChan = make(chan error)
	sigChan = make(chan os.Signal, 1)
)

func init() {
	flag.StringVar(&filename, "o", "", "local output file name")
	flag.DurationVar(&timeout, "T", 30*time.Minute, "timeout")
	flag.BoolVar(&insecure, "k", false, "do not verify the SSL certificate")
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//deal with signal
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go signalHandler(cancel)

	//deal with error
	go func() {
		err := <-errChan
		fmt.Printf("\n%v\n", err)
		fmt.Println(summary)
		os.Exit(1)
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
	outFile, err := parseFilename(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	if filename == "" {
		filename = outFile
	}

	ok, err := serverSupport(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !ok {
		fmt.Println("server not support for multiple request, you may download it with a web browser")
		return
	}

	summary = newResult(filesize)
	go func() {
		loader := newDownloader(ctx, filename, rawURL, summary)
		loader.do()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bar := newBar(summary)
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
	filesize = resp.ContentLength

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

func signalHandler(calcel context.CancelFunc) {
	<-sigChan
	calcel()
}

func usage() {
	fmt.Println(`
Usage: myaxel [options] url

optons:
  -T duration
        timeout (default 30m0s)
  -k    do not verify the SSL certificate
  -o string
        local output file name (default "default"
		`)
}
