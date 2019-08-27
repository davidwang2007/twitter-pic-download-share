package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestJson(t *testing.T) {

	f, err := os.OpenFile("D:\\ccwork\\video.json", os.O_RDONLY, 0644)
	//f, err := os.OpenFile("D:\\ccwork\\text.json", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("open file failed %v\n", err)
	}
	defer f.Close()
	//mp := make(map[string]interface{})
	raw := &TweetRaw{}
	err = json.NewDecoder(f).Decode(&raw)
	if err != nil {
		t.Fatalf("decode error %v\n", err)
	}
	//t.Logf("Got id %v\n", raw.GenTweetItem())

}
