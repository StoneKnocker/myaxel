package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	timeout time.Duration
	output  string
	ret     result
)

var (
	fileSize  int64
	startTime time.Time
	multiMod  = true
)

type result struct {
	downLen int64

	sync.Mutex
	f *os.File
}

func init() {
	flag.StringVar(&output, "o", "default", "local output file name")
	flag.DurationVar(&timeout, "T", 30*time.Minute, "timeout")
	startTime = time.Now()
	ret = result{}
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println(timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = ctx
	defer cancel()

	//TODO url check
	fileUrl := flag.Arg(0)

	var err error
	ret.f, err = os.Create(output)
	if err != nil {
		showError(err)
	}
	defer ret.f.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", fileUrl, nil)
	if err != nil {
		showError(err)
	}

	can, err := multiSupport(req)
	if err != nil {
		showError(err)
	}
	if can {
		multiDownlad(req, ret.f)
	} else {
		fmt.Println("not support or litle file, download in sigle thread...")
		singleDownload(req, ret.f)
	}

	summary()
}

func singleDownload(req *http.Request, f *os.File) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	n, err := f.Write(content)
	if err != nil {
		panic(err)
	}
	fmt.Printf("single done: %v", n)
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
			rangeStr := "bytes=" + strconv.FormatInt(rangeStart, 10) + "-" + strconv.FormatInt(rangeEnd, 10)
			fmt.Println("range:", rangeStr)
			newReq := cloneRequest(req)
			newReq.Header.Set("Range", rangeStr)

			resp, err := http.DefaultClient.Do(newReq)
			if err != nil {
				showError(err)
				return
			}
			defer resp.Body.Close()

			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				showError(err)
				return
			}
			ret.Lock()
			ret.f.Seek(rangeStart, os.SEEK_SET)
			ret.f.Write(content)
			ret.Unlock()

			atomic.AddInt64(&ret.downLen, int64(len(content)))
			fmt.Println("download: ", len(content))
		}(i)
	}
	wg.Wait()
}

func multiSupport(req *http.Request) (bool, error) {
	resp, err := http.Head(req.URL.String())
	if err != nil {
		return false, err
	}
	if resp.ContentLength < 1024 {
		return false, nil
	}
	fileSize = resp.ContentLength

	newReq := cloneRequest(req)
	newReq.Header.Set("Range", "Bytes=0-1")
	resp, err = http.DefaultClient.Do(newReq)
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
	// fmt.Println(err)
	panic(err)
	os.Exit(1)
}

func summary() {
	timeSpent := time.Since(startTime)
	fmt.Printf("fileSize: %d, download %d in %v, %.2f bytes/s", fileSize, ret.downLen, timeSpent, float64(fileSize)/timeSpent.Seconds())
}
