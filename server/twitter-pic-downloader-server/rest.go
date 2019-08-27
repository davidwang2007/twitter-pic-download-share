package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"google.golang.org/appengine/urlfetch"
)

//const URL_REQUEST_TOKEN = "https://www.baidu.com/twitter/oauth/request_token"
const URL_REQUEST_TOKEN = "https://api.twitter.com/oauth/request_token"
const URL_AUTHENTICATE = "https://api.twitter.com/oauth/authenticate"
const URL_ACCESS_TOKEN = "https://api.twitter.com/oauth/access_token"
const URL_TWEET_SHOW = "https://api.twitter.com/1.1/statuses/show.json"

//HandleOAuth handle browser /twitter/oauth request
//generate oauth_token by the consumer_key&consumer_secret
//then send redirect
//handle "/twitter/oauth"
func HandleOAuth(w http.ResponseWriter, req *http.Request) {
	//get the oauth_token
	var params = map[string]string{"x_auth_access_type": "read"}
	var ctx = appengine.NewContext(req)

	msg, err := doTwitterRequest(req, "POST", myOAuthToken, myOAuthSecret, URL_REQUEST_TOKEN, params)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	values, err := url.ParseQuery(msg)
	if err != nil {
		log.Errorf(ctx, "When parse response to url.Values from api.twitter %s got error %v\n", URL_REQUEST_TOKEN, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tk := values.Get("oauth_token")
	ts := values.Get("oauth_token_secret")
	//store
	SaveTempRecord(ctx, tk, ts)
	http.Redirect(w, req, fmt.Sprintf("%s?oauth_token=%s", URL_AUTHENTICATE, tk), http.StatusFound)

}

//HandleOAuthCallback handle twitter redirect
// "/twitter/oauth/callback?oauth_token=xx&oauth_verifier=xxx"
//此处的oauth_token仍然是临时的
func HandleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	var ctx = appengine.NewContext(req)
	ot := req.FormValue("oauth_token")
	ov := req.FormValue("oauth_verifier")
	if ot == "" || ov == "" {
		log.Warningf(ctx, "/twitter/oauth/callback got bad request, parameter not suitable\n")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := VerifyTempRecord(ctx, ot, ov); err != nil {
		log.Errorf(ctx, "VerifyTempRecord calling failed! %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//exchange the final oauth_token&secret
	tmpRecord, err := GetTempRecord(ctx, ot)
	if err != nil {
		log.Errorf(ctx, "Get OAuthTempRecord by token %s failed! %v\n", ot, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if tmpRecord == nil {
		log.Warningf(ctx, "OAuthTempRecord Not Found by token %s \n", ot)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var params = map[string]string{"oauth_verifier": ov}
	msg, err := doTwitterRequest(req, "POST", ot, tmpRecord.OAuthSecret, URL_ACCESS_TOKEN, params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	values, err := url.ParseQuery(msg)
	if err != nil {
		log.Warningf(ctx, "When parse response to url.Values from api.twitter %s got error %v\n", URL_ACCESS_TOKEN, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tk := values.Get("oauth_token")
	ts := values.Get("oauth_token_secret")
	userID := values.Get("user_id")
	screenName := values.Get("screen_name")
	if err := SaveRecord(ctx, tk, ts, userID, screenName); err != nil {
		log.Errorf(ctx, "Save record failed %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//send redirect
	//带req参数的话就不用再加上scheme之类的了
	//http.Redirect(w, req, fmt.Sprintf("%s://%s/twitter/oauth/android?oauth_token=%s", req.URL.Scheme, req.URL.Host, tk), http.StatusFound)
	http.Redirect(w, req, fmt.Sprintf("/twitter/oauth/android?oauth_token=%s", tk), http.StatusFound)

}

//HandleOAuthAndroid handle oauth android
// "/twitter/oauth/android"
func HandleOAuthAndroid(w http.ResponseWriter, req *http.Request) {

	var ctx = appengine.NewContext(req)
	var tk = req.FormValue("oauth_token")
	if tk == "" {
		log.Warningf(ctx, "/twitter/oauth/android got bad request, parameter not suitable\n")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	record, err := GetRecord(ctx, tk)

	if err != nil {
		log.Warningf(ctx, "Get OAuthRecord by token %s failed! %v\n", tk, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if record == nil {
		log.Warningf(ctx, "OAuthRecord Not Found by token %s \n", tk)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	//
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("oauth succeed"))

}

//HandleGetTweet handle show tweet
// "/twitter/show"
func HandleGetTweet(w http.ResponseWriter, req *http.Request) {

	var ctx = appengine.NewContext(req)
	var tk = req.FormValue("oauth_token")
	var tweetID = req.FormValue("tweet_id")
	if tk == "" || tweetID == "" {
		log.Warningf(ctx, "/twitter/oauth/android got bad request, parameter not suitable\n")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	record, err := GetRecord(ctx, tk)

	if err != nil {
		log.Errorf(ctx, "Get OAuthRecord by token %s failed! %v\n", tk, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if record == nil {
		log.Warningf(ctx, "OAuthRecord Not Found by token %s \n", tk)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//先查本地
	tid, _ := strconv.ParseInt(tweetID, 10, 64)
	tweetItem, err := GetTweet(ctx, tid)
	if err != nil {
		log.Errorf(ctx, "GetTweet from storage by %s failed! %v\n", tweetID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if tweetItem != nil {
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		tweetItem.EncodeJSON(w)
		return
	}

	//make request to api.twitter :show
	var params = map[string]string{
		"id":                   tweetID,
		"trim_user":            "true",
		"include_my_retweet":   "false",
		"include_card_uri":     "false",
		"include_ext_alt_text": "false",
		"include_entities":     "false",
	}
	msg, err := doTwitterRequest(req, "GET", record.OAuthToken, record.OAuthSecret, URL_TWEET_SHOW, params)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	raw := &TweetRaw{}
	err = json.NewDecoder(strings.NewReader(msg)).Decode(raw)
	if err != nil {
		log.Errorf(ctx, "Decode tweet:show failed! %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tweetItem = raw.GenTweetItem(ctx)
	if err = SaveTweetItem(ctx, tweetItem); err != nil {
		log.Errorf(ctx, "save tweet failed! %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	tweetItem.EncodeJSON(w)
}

//doTwitterRequest do api.twitter rest request
func doTwitterRequest(req *http.Request, method, oauthToken, oauthSecret, url string, params map[string]string) (string, error) {
	var ctx = appengine.NewContext(req)
	var headers = InitHeader(consumerKey, oauthToken)

	var signature = GenSignature(method, url, consumerSecret, oauthSecret, params, headers)
	var oauthHeader = GenOAuthHeader(headers, signature)

	log.Debugf(ctx, "consumerKey: %s\n", consumerKey)
	log.Debugf(ctx, "consumerSecret: %s\n", consumerSecret)
	log.Debugf(ctx, "oauthToken: %s\n", oauthToken)
	log.Debugf(ctx, "oauthSecret: %s\n", oauthSecret)
	log.Debugf(ctx, "oauthHeader: %s\n", oauthHeader)
	log.Debugf(ctx, "url: %s %s\n", method, url)
	log.Debugf(ctx, "params: %v\n", params)

	/*
		if 2 > 1 {
			return "", errors.New("MOCK")
		}
	*/

	var reqTwitter *http.Request
	var err error
	if method != "GET" {
		reqTwitter, err = http.NewRequest(method, url, strings.NewReader(URLEncodeParams(params)))
	} else {
		if strings.Index(url, "?") >= 0 {
			reqTwitter, err = http.NewRequest(method, fmt.Sprintf("%s&%s", url, URLEncodeParams(params)), nil)
		} else {
			reqTwitter, err = http.NewRequest(method, fmt.Sprintf("%s?%s", url, URLEncodeParams(params)), nil)
		}
	}
	if err != nil {
		log.Errorf(ctx, "When init request %s got error %v\n", url, err)
		return "", err
	}
	reqTwitter.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqTwitter.Header.Set("Authorization", oauthHeader)
	//remember request with appengine urlfetch client
	var client = client(ctx) //does not work

	//resp, err := http.DefaultClient.Do(reqTwitter.WithContext(ctx))
	//resp, err := client.Do(reqTwitter.WithContext(ctx))
	resp, err := client.Do(reqTwitter)

	if err != nil {
		log.Errorf(ctx, "When request %s got error %v\n", url, err)
		return "", err
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "When read response from api.twitter %s got error %v\n", url, err)
		return "", err
	}

	if resp.StatusCode/100 != 2 {
		log.Debugf(ctx, "Request twitter[%s] got response %v\n", url, string(bs))
		return "", fmt.Errorf("Unexpected status code %d", resp.StatusCode)
	}

	return string(bs), nil
}

// create a custom error to know if a redirect happened
var RedirectAttemptedError = errors.New("redirect")

func noCheckRedirect(req *http.Request, via []*http.Request) error {
	return RedirectAttemptedError
}

var DefaultTransport http.RoundTripper = &http.Transport{Proxy: http.ProxyFromEnvironment}

func client(ctx context.Context) *http.Client {
	var client = urlfetch.Client(ctx)
	client.CheckRedirect = noCheckRedirect
	if appengine.IsDevAppServer() {
		//return http.DefaultClient
		log.Debugf(ctx, "debug mode, set proxy for transport\n")
		client.Transport = DefaultTransport
	}
	return client
}
