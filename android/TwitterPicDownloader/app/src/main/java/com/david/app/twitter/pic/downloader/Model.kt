package com.david.app.twitter.pic.downloader

object Model{
    data class TweetItem(val id:Long, val text:String, val tweetTime:String, val createTime:String, val type:String, val mediaURL:String, val videoURL:String)
}