package com.david.app.twitter.pic.downloader

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.support.design.widget.Snackbar
import android.support.v7.app.AlertDialog
import android.util.Log
import android.view.Menu
import android.view.MenuItem
import android.view.View
import android.widget.Toast
import kotlinx.android.synthetic.main.activity_share.*
import java.io.File

class TwitterShareActivity : BaseActivity(), TweetItemListener {

    //回收
    var recycleBitMap: Bitmap? = null
    //如果是视频缩略图的话，退出后要删除掉这个多余的
    var readyDeleted: File? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_share)

        initAd(adView)

        Utils.verifyStoragePermissions(this)

        //先隐藏不必要的，只显示滚动条,转圈条
        listOf<View>(cardView, imageView, progressBarH, chip,downloadBtn).forEach { it.visibility = View.GONE }

        intent?.getStringExtra(Intent.EXTRA_TEXT)?.apply {
            val index = this.lastIndexOf("https://")
            val uri = Uri.parse(if(index >= 0) this.substring(index) else "https://davidwang.site")
            Log.i(TAG, "got uri $uri")
            Log.i(TAG, "host: ${uri.host}")
            Log.i(TAG, "path: ${uri.path}")

            if (uri.host != "twitter.com" || uri.pathSegments.size != 3) {
                AlertDialog.Builder(this@TwitterShareActivity)
                        .setTitle(R.string.dialog_title_prompt)
                        .setMessage(R.string.invalid_url)
                        .setPositiveButton(R.string.ok) { _, _ -> finish() }
                        .show()

                return
            }

            Utils.prefEdit(this@TwitterShareActivity).putString("tweet_id", uri.pathSegments.last()).apply()
        }

        //如果第一次未授权，则提示授权
        if (Utils.pref(this).getString("oauth_token", "") == "") {
            val builder = AlertDialog.Builder(this)
            builder.setTitle(R.string.dialog_title_prompt)
            builder.setMessage(R.string.oauth_needed_first_time)
            builder.setPositiveButton(R.string.ok) { _, _ ->
                startActivity(Intent(this, OAuthActivity::class.java))
                finish()
            }
            builder.setNegativeButton(R.string.cancel) { _, _ ->
                finish()
            }
            builder.show()
            return
        }

        if(Utils.pref(this).getString("tweet_id","") == ""){
            AlertDialog.Builder(this@TwitterShareActivity)
                    .setTitle(R.string.dialog_title_prompt)
                    .setMessage(R.string.invalid_url)
                    .setPositiveButton(R.string.ok) { _, _ -> finish() }
                    .show()
            return
        }
        FetchTweetItemTask(this, this).execute()


        textView.setOnLongClickListener {
            val clipMgmr = this.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            val clipData = ClipData.newPlainText("text",textView.text)
            clipMgmr.primaryClip = clipData
            Toast.makeText(this,R.string.copy_2,Toast.LENGTH_SHORT).show()
            true
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        //imageView.setImageResource(null)
        recycleBitMap?.recycle()
        readyDeleted?.delete()

    }

    override fun onCreateOptionsMenu(menu: Menu): Boolean {
        menuInflater.inflate(R.menu.menu_close, menu)
        return true
    }

    override fun onOptionsItemSelected(item: MenuItem) = when (item.itemId) {
        R.id.menu_close -> {
            finish();true
        }
        else -> super.onOptionsItemSelected(item)
    }

    override fun ok(item: Model.TweetItem) {
        //移除临时存下的id
        Utils.prefEdit(this).remove("tweet_id").apply()
        cardView.visibility = View.VISIBLE
        textView.text = item.text
        if (item.type == "text") {
            AlertDialog.Builder(this)
                    .setTitle(R.string.dialog_title_prompt)
                    .setMessage(R.string.no_media_found)
                    .setPositiveButton(R.string.ok, { _, _ -> finish() })
                    .show()
            return
        }
        //下载缩略图或真图
        downloadMedia(item.mediaURL, { fileName ->
            progressBar.visibility = View.GONE
            val dir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
            var targetFile = dir.resolve(fileName)
            val bp = BitmapFactory.decodeFile(targetFile.absolutePath)
            imageView.setImageBitmap(bp)
            imageView.visibility = View.VISIBLE
            recycleBitMap = bp
            //bp.recycle()
            if (item.videoURL == "") {
                //Snackbar.make(imageView, "Save to \"$fileName\"", Snackbar.LENGTH_LONG).show()
                chip.visibility = View.VISIBLE
                chip.chipText = "SAVE AS \"$fileName\""
            }else{
                //下载视频文件
                downloadMedia(item.videoURL,{videoName ->
                    readyDeleted = targetFile
                    targetFile = dir.resolve(fileName)
                    //Snackbar.make(imageView, "Save to \"$videoName\"", Snackbar.LENGTH_LONG).show()
                    chip.visibility = View.VISIBLE
                    chip.chipText = "SAVE AS \"$videoName\""
                })
            }
            //最后展示发送分享按钮
            downloadBtn.visibility = View.VISIBLE
            downloadBtn.setOnClickListener {
                Utils.sendMedia(this,targetFile)
            }
        })


    }

    fun downloadMedia(url: String, successCallback: (fname: String) -> Unit) {
        MediaDownloadTask(this, object : MediaDownloadListener {
            override fun ok(fileName: String) {
                progressBarH.visibility = View.GONE
                successCallback(fileName)
            }

            override fun fail(msg: String) {
                progressBarH.visibility = View.GONE
                Snackbar.make(textView, "ERROR: $msg", Snackbar.LENGTH_LONG).show()
                downloadBtn.visibility = View.VISIBLE
                downloadBtn.setText(R.string.close)
                downloadBtn.setOnClickListener { finish() }
            }

            override fun progress(current: Int, total: Int) {
                if (progressBarH.visibility == View.GONE) {
                    progressBarH.visibility = View.VISIBLE
                    progressBarH.max = total
                }
                progressBarH.progress = current

            }
        }).execute(url)

    }

    override fun fail(msg: String) {
        progressBar.visibility = View.GONE
        textView.visibility = View.VISIBLE
        textView.text = msg
        Snackbar.make(textView, msg, Snackbar.LENGTH_LONG).show()
        val is400 = msg == "STATUS CODE: 400" //表示要重新鉴权
        AlertDialog.Builder(this)
                .setTitle(R.string.dialog_title_prompt)
                .setMessage(if(is400) getString(R.string.invalid_oauth_token) else msg)
                .setPositiveButton(R.string.ok, { _, _ ->
                    if(is400) startActivity(Intent(this,OAuthActivity::class.java))
                    finish()
                })
                .show()
        return
    }
}
