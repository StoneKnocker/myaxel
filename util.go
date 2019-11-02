package main

import (
	"net/url"
	"path/filepath"
	"strings"
)

func parseFilename(rawURL string) (string, error) {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}
	filename := filepath.Base(rawURL)
	if n := strings.Index(filename, "?"); n != -1 {
		filename = filename[:n]
	}
	return filename, nil
}
