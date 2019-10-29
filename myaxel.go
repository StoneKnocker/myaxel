package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var output string
var timeout time.Duration

func init() {
	flag.StringVar(&output, "o", "default", "local output file name")
	flag.DurationVar(&timeout, "T", 30*time.Minute, "timeout")
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = ctx
	defer cancel()

	//TODO url check
	fileUrl := flag.Arg(0)

	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", fileUrl, nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(content)
	if err != nil {
		panic(err)
	}
}
