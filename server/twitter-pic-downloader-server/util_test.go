package main

import "testing"

//TestGenSignature
func TestGenSignature(t *testing.T) {

	params := map[string]string{
		"x_auth_access_type": "read",
		//"oauth_callback":     "https://www.davidwang.site",
	}
	consumeKey := "WqMhGBZr7MkjqPiN1qIyqOKUq"
	oauthToken := "414627346-1ctEs7wQ96aB2B20Wm8CPWFRcfsfSrGNDvof3Y0H"
	headers := InitHeader(consumeKey, oauthToken)
	headers["oauth_nonce"] = "DX7W7FuYySktgDjWoDoxPNuk7jNZy0KWEn1JxU"
	headers["oauth_timestamp"] = "1527421496"
	sign := GenSignature("POST", "https://api.twitter.com/oauth/request_token", "XHM6iO8qh7aZpQfQg18MEU993dbav7N4JYgBf8UWA2Yny8lT9Q", "5HdPyfbJQS07ay0Sj2v9qGCeRkpuhfZcOnc0P2WdtPGpw", params, headers)
	t.Logf("signature is %s\n", sign)

	var oauthHeader = GenOAuthHeader(headers, sign)
	t.Logf("OAuth header is %s\n", oauthHeader)

}
