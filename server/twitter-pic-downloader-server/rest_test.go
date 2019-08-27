package main

import (
	"net/url"
	"testing"
)

func TestURLEncode(t *testing.T) {
	var vs = url.Values{}
	vs.Add("Now K", "Yes 我 is ? & k")
	vs.Add("b", "ok")
	t.Logf("after url encode: %s\n", vs.Encode())

	vs, err := url.ParseQuery("a%20b=you%E6%88%91%E4%BA%86")
	if err != nil {
		t.Fatalf("Got error %v\n", err)
	}
	t.Logf("Got params %v\n", vs)

}
