package main

import (
	"strconv"
	"testing"
)

func TestParseFilename(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{
			"http://example.com/a.txt",
			"a.txt",
		},
		{
			"http://example.com/a.txt?a=b",
			"a.txt",
		},
	}
	for idx, tC := range testCases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			if ret, err := parseFilename(tC.url); err != nil || ret != tC.expected {
				if err != nil {
					t.Error(err)
					return
				}
				t.Errorf("parseFilename(%s), expected: %s, actual: %s", tC.url, tC.expected, ret)
			}
		})
	}
}
