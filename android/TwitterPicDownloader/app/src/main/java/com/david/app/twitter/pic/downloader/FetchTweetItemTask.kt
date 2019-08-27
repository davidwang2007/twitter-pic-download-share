package com.david.app.twitter.pic.downloader

import android.content.Context
import android.os.AsyncTask
import android.util.Log
import com.google.gson.Gson
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL

class FetchTweetItemTask(val ctx: Context,val listener: TweetItemListener): AsyncTask<Void, String, Model.TweetItem>(){
    override fun onPostExecute(result: Model.TweetItem?) {
        if(result != null ) listener.ok(result)
    }

    override fun onProgressUpdate(vararg values: String) {
        //Snackbar.make(textView,values[0],Snackbar.LENGTH_LONG).show()
        listener.fail(values.joinToString())
    }

    override fun doInBackground(vararg params: Void?): Model.TweetItem? {
        val oauthToken = Utils.pref(ctx).getString("oauth_token","")
        val tweetId = Utils.pref(ctx).getString("tweet_id","")
        val conn = URL("https://${ctx.getString(R.string.base_url)}/twitter/show?oauth_token=$oauthToken&tweet_id=$tweetId").openConnection() as HttpURLConnection
        try{
            conn.connect()
            if(conn.responseCode / 100 != 2) throw IOException("STATUS CODE: ${conn.responseCode}")
            return Gson().fromJson(conn.inputStream)
        }catch(ex:IOException){
            Log.e(BaseActivity.TAG,"fetch tweet get error: ${ex.message}",ex)
            publishProgress(ex.message)
            return null
        }finally{
            try{conn.inputStream.close()}catch(ex:IOException){/*ignore*/}
        }

    }

}
interface TweetItemListener{
    fun ok(item:Model.TweetItem)
    fun fail(msg:String)
}

