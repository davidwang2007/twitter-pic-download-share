package main

//主

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

//Nonce generator
//we should strip out all non-word characters
var nonWordReg = regexp.MustCompile("[^\\w]")

//Nonce create Nonce
func Nonce() string {
	bs := make([]byte, 32) //32bytes
	rand.Read(bs)
	line := base64.StdEncoding.EncodeToString(bs)
	return nonWordReg.ReplaceAllString(line, "")
}

//BaseURL get base url
func BaseURL(addr string) string {
	base := addr
	i := strings.Index(addr, "?")
	if i > -1 {
		base = addr[:i]
	}
	return base
}

//GenSignature generate signature
func GenSignature(method, url, consumerSecret, tokenSecret string, params, headers map[string]string) string {

	//put all the headers and params into a map
	allParams := make(map[string]string)
	//put all keys into keys, for sort
	keys := make([]string, 0, len(params)+len(headers))
	for k, v := range params {
		allParams[k] = v
		keys = append(keys, k)
	}
	for k, v := range headers {
		allParams[k] = v
		keys = append(keys, k)
	}

	//must sort according to the alphabet
	sort.Strings(keys)
	var headString string
	for i, k := range keys {
		if i > 0 {
			headString += "&"
		}
		headString += fmt.Sprintf("%s=%s", URLEncode(k), URLEncode(allParams[k]))
	}
	var signatureBaseString = fmt.Sprintf("%s&%s&%s", URLEncode(method), URLEncode(BaseURL(url)), URLEncode(headString))

	k := fmt.Sprintf("%s&%s", consumerSecret, tokenSecret)
	mac := hmac.New(sha1.New, []byte(k))
	mac.Write([]byte(signatureBaseString))
	bs := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(bs)
}

var urlPlusRegExp = regexp.MustCompile("\\+")

// URLEncode encodes a string like Javascript's encodeURIComponent()
//EncodeParams little like url.Values.Encode()
//http://stackoverflow.com/questions/13820280/encode-decode-urls
//由于url.QueryEscape将空格转成了+而不是%20，所以要使用这一个
func URLEncode(str string) string {
	line := url.QueryEscape(str)
	return urlPlusRegExp.ReplaceAllString(line, "%20")
}

//URLEncodeParams url encode params,
//we did not use the url.Values.Encode()
//because it not encode '+' as %20
func URLEncodeParams(params map[string]string) string {
	var line string
	for k, v := range params {
		line += fmt.Sprintf("%s=%s&", URLEncode(k), URLEncode(v))
	}

	//delete the tail '&'
	return line[:len(line)-1]
}

//FormatDate format created_at to yyyy-MM-dd HH:mm:ss
func FormatDate(createdAt string) string {
	ti, err := time.Parse(time.RubyDate, createdAt)
	if err != nil {
		return createdAt
	}
	ti = ti.Add(time.Hour * 8) //时差8小时
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", ti.Year(), ti.Month(), ti.Day(), ti.Hour(), ti.Minute(), ti.Second())
}

//InitHeader init twitter api http request header
func InitHeader(consumerKey, oauthToken string) (mp map[string]string) {
	mp = make(map[string]string)
	mp["oauth_consumer_key"] = consumerKey
	mp["oauth_nonce"] = Nonce()
	//mp["oauth_nonce"] = "abcdefg"
	mp["oauth_timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	//mp["oauth_timestamp"] = "1234567890"
	mp["oauth_signature_method"] = "HMAC-SHA1"
	mp["oauth_token"] = oauthToken
	mp["oauth_version"] = "1.0"
	return
}

//GenOAuthHeader generate Authorization: OAuth header
func GenOAuthHeader(headers map[string]string, signature string) string {
	var line = "OAuth "
	for k, v := range headers {
		line += fmt.Sprintf("%s=\"%s\", ", URLEncode(k), URLEncode(v))
	}

	line += fmt.Sprintf("oauth_signature=\"%s\"", URLEncode(signature))

	return line
}
