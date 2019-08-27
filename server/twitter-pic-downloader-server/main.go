package main

import (
	"net/http"
	"os"

	"google.golang.org/appengine"
)

//remember set consumer key&secret to app.yaml environment
var consumerKey = os.Getenv("CONSUMER_KEY")
var consumerSecret = os.Getenv("CONSUMER_SECRET")
var myOAuthToken = os.Getenv("OAUTH_TOKEN")
var myOAuthSecret = os.Getenv("OAUTH_SECRET")

func main() {

	http.HandleFunc("/twitter/oauth/android", HandleOAuthAndroid)
	http.HandleFunc("/twitter/oauth/callback", HandleOAuthCallback)
	http.HandleFunc("/twitter/show", HandleGetTweet)
	http.HandleFunc("/twitter/oauth", HandleOAuth)
	http.Handle("/", http.NotFoundHandler())

	appengine.Main()

}
