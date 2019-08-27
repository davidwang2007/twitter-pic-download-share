package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//OAuthTempRecord temp record
//the temp record, just for exchange for real oauth_token
type OAuthTempRecord struct {
	OAuthToken    string    //oauth_token
	OAuthSecret   string    //oauth_secret
	OAuthVerifier string    //oauth_verifier when the user login succeed
	InitTime      time.Time //the time when OAuthToken created
	VerifyTime    time.Time //the time when user login succeed
}

//OAuthRecord define some struct
//the final effective record
type OAuthRecord struct {
	OAuthToken  string    //oauth_token
	OAuthSecret string    //oauth_secret
	UserID      string    //user id
	ScreenName  string    //screen name
	CreateTime  time.Time //create time
}

//TweetItem tweet item
type TweetItem struct {
	ID         int64     `json:"id"`
	Text       string    `json:"text"`
	TweetTime  time.Time `json:"tweetTime"`
	CreateTime time.Time `json:"createTime"` //tweet create time
	Type       string    `json:"type"`       //txt, photo, animated_gif,video
	MediaURL   string    `json:"mediaURL"`   // just store the https url, the thumbnail or the orginal for jpg, 即截图或真实文件
	VideoURL   string    `json:"videoURL"`   //for gif and video 针对video, 存储比特率最高的那一个，针对gif也是一个mp4
}

//EncodeJSON encode the tweetitem to json
func (item *TweetItem) EncodeJSON(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(item)
}

//TweetRaw  raw tweet show
type TweetRaw struct {
	ID               int64                               `json:"id"`
	CreatedAt        string                              `json:"created_at"`
	Text             string                              `json:"text"`
	Truncated        bool                                `json:"truncated"`
	ExtendedEntities map[string][]map[string]interface{} `json:"extended_entities"` //just take the media
}

//GenTweetItem generate tweet item
func (raw *TweetRaw) GenTweetItem(ctx context.Context) (item *TweetItem) {
	item = &TweetItem{}
	item.ID = raw.ID
	item.CreateTime = time.Now()
	item.TweetTime, _ = time.Parse(time.RubyDate, raw.CreatedAt)
	item.Text = raw.Text
	if raw.ExtendedEntities == nil || raw.ExtendedEntities["media"] == nil {
		item.Type = "text"
		return
	}
	medias := raw.ExtendedEntities["media"]
	if len(medias) == 0 {
		item.Type = "text"
		return
	}
	if len(medias) > 1 {
		log.Warningf(ctx, "medias has %d items, we just take the first one\n", len(medias))
	}
	media := medias[0]
	item.MediaURL = media["media_url_https"].(string)
	item.Type = media["type"].(string)
	videoInfo, ok := media["video_info"].(map[string]interface{})
	if !ok { //表明是photo
		return
	}
	variants, ok := videoInfo["variants"].([]interface{})
	if !ok { //表明是photo
		log.Errorf(ctx, "Expected variants, but no\n")
		return
	}
	if len(variants) == 0 {
		log.Warningf(ctx, "variants length should larget than 0!\n")
		return
	}
	var bitrate int64
	var url string
	for _, variant := range variants {
		mp := variant.(map[string]interface{})
		f, _ := mp["bitrate"].(float64)
		br := int64(f)
		if br >= bitrate {
			bitrate = br
			url = mp["url"].(string)
		}
	}
	item.VideoURL = url

	return
}

//SaveTempRecord save OAuthTempRecord
//save to gcloud storage
func SaveTempRecord(ctx context.Context, oauthToken, oauthSecret string) error {
	record := &OAuthTempRecord{}
	record.OAuthToken = oauthToken
	record.OAuthSecret = oauthSecret
	record.InitTime = time.Now()
	//var key = datastore.NewIncompleteKey(ctx, "OAuthTempRecord", nil)
	var key = datastore.NewKey(ctx, "OAuthTempRecord", oauthToken, 0, nil)
	key, err := datastore.Put(ctx, key, record)
	log.Debugf(ctx, "Now key is %v\n", key)
	return err
}

//VerifyTempRecord verify temp record
func VerifyTempRecord(ctx context.Context, oauthToken, oauthVerifier string) error {
	//first query out
	//second set oauthVerifier then update it
	var key = datastore.NewKey(ctx, "OAuthTempRecord", oauthToken, 0, nil)
	q := datastore.NewQuery("OAuthTempRecord").Filter("__key__ =", key).Limit(1)

	t := q.Run(ctx)
	var record = &OAuthTempRecord{}
	_, err := t.Next(record)
	if err == datastore.Done || err == datastore.ErrNoSuchEntity { //证明没有这个记录
		return fmt.Errorf("OAuthTempRecord[%s] not found", oauthToken)
	}

	record.VerifyTime = time.Now()
	record.OAuthVerifier = oauthVerifier
	_, err = datastore.Put(ctx, key, record)
	return err
}

//GetTempRecord get OAuthTempRecord from storage
func GetTempRecord(ctx context.Context, oauthToken string) (record *OAuthTempRecord, err error) {
	var key = datastore.NewKey(ctx, "OAuthTempRecord", oauthToken, 0, nil)
	q := datastore.NewQuery("OAuthTempRecord").Filter("__key__ =", key).Limit(1)

	t := q.Run(ctx)
	record = &OAuthTempRecord{}
	_, err = t.Next(record)
	if err == datastore.Done || err == datastore.ErrNoSuchEntity { //证明没有这个记录
		return nil, fmt.Errorf("OAuthTempRecord[%s] not found", oauthToken)
	} else if err != nil {
		return nil, err
	}

	return record, nil
}

//SaveRecord save OAuthRecord
func SaveRecord(ctx context.Context, oauthToken, oauthSecret, userID, screenName string) error {
	var record = &OAuthRecord{
		oauthToken, oauthSecret, userID, screenName, time.Now(),
	}
	var key = datastore.NewKey(ctx, "OAuthRecord", oauthToken, 0, nil)
	key, err := datastore.Put(ctx, key, record)
	return err
}

//GetRecord get OAuthRecord from storage
func GetRecord(ctx context.Context, oauthToken string) (record *OAuthRecord, err error) {
	var key = datastore.NewKey(ctx, "OAuthRecord", oauthToken, 0, nil)
	q := datastore.NewQuery("OAuthRecord").Filter("__key__ =", key).Limit(1)

	t := q.Run(ctx)
	record = &OAuthRecord{}
	_, err = t.Next(record)
	if err == datastore.Done || err == datastore.ErrNoSuchEntity { //证明没有这个记录
		//return nil, fmt.Errorf("OAuthRecord[%s] not found", oauthToken)
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return record, nil
}

//SaveTweetItem save tweet item
func SaveTweetItem(ctx context.Context, item *TweetItem) error {
	var key = datastore.NewKey(ctx, "TweetItem", "", item.ID, nil)
	key, err := datastore.Put(ctx, key, item)
	return err
}

//GetTweet find tweet from storage
func GetTweet(ctx context.Context, id int64) (tweet *TweetItem, err error) {
	var key = datastore.NewKey(ctx, "TweetItem", "", id, nil)
	q := datastore.NewQuery("TweetItem").Filter("__key__ =", key).Limit(1)

	t := q.Run(ctx)
	tweet = &TweetItem{}
	_, err = t.Next(tweet)
	if err == datastore.Done || err == datastore.ErrNoSuchEntity { //证明没有这个记录
		//return nil, fmt.Errorf("TweetItem[%d] not found", id)
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return tweet, nil
}
