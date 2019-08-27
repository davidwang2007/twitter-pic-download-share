package com.david.app.twitter.pic.downloader

import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.view.Menu
import android.view.MenuItem
import android.webkit.JavascriptInterface
import android.webkit.WebView
import android.webkit.WebViewClient
import kotlinx.android.synthetic.main.activity_oauth.*

class OAuthActivity : BaseActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_oauth)

        initAd(adView)

        webView.settings.javaScriptEnabled = true
        webView.addJavascriptInterface(JavaScriptInterface(), JAVASCRIPT_OBJ)
        webView.webViewClient = object: WebViewClient(){
            override fun onPageFinished(view: WebView?, url: String?) {
                injectJavaScriptFunction()
                val uri = Uri.parse(url)
                if(uri.path == "/twitter/oauth/android") {
                    val token = uri.getQueryParameter("oauth_token")
                    Utils.prefEdit(this@OAuthActivity).putString("oauth_token",token).apply()
                    startActivity(Intent(this@OAuthActivity,TwitterShareActivity::class.java))
                    finish()
                }
            }
        }
        val uri = "https://${getString(R.string.base_url)}/twitter/oauth"
        webView.loadUrl(uri)


    }

    override fun onCreateOptionsMenu(menu: Menu): Boolean {
        menuInflater.inflate(R.menu.menu_close,menu)
        return true
    }

    override fun onOptionsItemSelected(item: MenuItem) = when(item.itemId){
            R.id.menu_close -> {finish();true}
            else -> super.onOptionsItemSelected(item)
        }

    override fun onDestroy() {
        webView.removeJavascriptInterface(JAVASCRIPT_OBJ)
        super.onDestroy()
    }

    private fun injectJavaScriptFunction(){
        webView.loadUrl("javascript: " +
                "window.androidObj.textToAndroid = function(message) { " +
                JAVASCRIPT_OBJ + ".oauthCallback(message) }")

    }

    //https://medium.com/@elye.project/making-android-interacting-with-web-app-921be14f99d8
    private inner class JavaScriptInterface{
        @JavascriptInterface
        fun oauthCallback(userId:String){
            getSharedPreferences("settings", Context.MODE_PRIVATE).edit().putString("userid",userId).commit()
        }
    }

    companion object {
        private val JAVASCRIPT_OBJ = "javascript_obj"
        private val BASE_URL = "file:///android_asset/webview.html"
    }
}
