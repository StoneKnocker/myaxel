package main

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
	timeout    time.Duration
	filename   string
	insecure   bool
	routineNum int

	httpClient = http.DefaultClient

	errChan = make(chan error)
	sigChan = make(chan os.Signal, 1)
)

func init() {
	flag.StringVar(&filename, "o", "", "local output file name")
	flag.DurationVar(&timeout, "T", 30*time.Minute, "timeout")
	flag.BoolVar(&insecure, "k", false, "do not verify the SSL certificate")
	flag.IntVar(&routineNum, "n", 2, "specify an alternative number of connections")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//deal with signal
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go signalHandler(cancel)

	//deal with error
	var summary *Summary
	go func() {
		for {
			select {
			case err := <-errChan:
				if err != nil {
					fmt.Printf("\n%v\n", err)
					fmt.Println(summary)
					os.Exit(1)
				}
			default:
			}
		}
	}()

	rawURL := flag.Arg(0)
	outFile, err := parseFilename(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	if filename == "" {
		filename = outFile
	}

	filesize, err := fetchFilesize(rawURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	if filesize == 0 {
		fmt.Println("server not support for multiple request, you may download it with a web browser")
		return
	}

	summary = NewSummary(filesize)
	go func() {
		loader := NewDownloader(ctx, routineNum, filename, rawURL, summary)
		loader.Do()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		bar := NewBar(summary)
		bar.Show()
	}()
	wg.Wait()

	fmt.Println(summary)
}

func fetchFilesize(rawURL string) (int64, error) {
	var filesize int64
	resp, err := http.Head(rawURL)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, errors.New("can't fetch the file")
	}
	if resp.Header.Get("Content-Length") == "" {
		return 0, errors.New("can't get the file length")
	}
	filesize = resp.ContentLength

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", "Bytes=0-1")
	resp, err = httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 206 {
		return 0, nil
	}
	return filesize, nil
}

func signalHandler(calcel context.CancelFunc) {
	<-sigChan
	calcel()
}

func errCollecter(err error) {
	errChan <- err
}

func usage() {
	fmt.Println(`
Usage: myaxel [options] url

optons:
	-T duration
		timeout (default 30m0s)
	-k	do not verify the SSL certificate
	-o string
		local output file name (default "")
	-n int
		specify an alternative number of connections (default 2)
		`)
}
