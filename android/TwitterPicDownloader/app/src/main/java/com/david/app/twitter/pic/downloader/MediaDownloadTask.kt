package com.david.app.twitter.pic.downloader

import android.content.Context
import android.net.Uri
import android.os.AsyncTask
import android.os.Environment
import android.util.Log
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL

/**
 * 媒体文件下载器
 * 不论是图片还是视频都下载至本地Download文件夹
 */
class MediaDownloadTask(val ctx: Context, val listener: MediaDownloadListener):AsyncTask<String,String,String>() {
    override fun onPostExecute(result: String) {
        if(result.startsWith("error:")) listener.fail(result.substringAfter(":"))
        else listener.ok(result)
    }

    override fun onProgressUpdate(vararg values: String) {
         listener.progress(values.first().toInt(),values.last().toInt())
    }

    override fun doInBackground(vararg params: String): String {
        val mediaUri = Uri.parse(params.first())
        val fileName = mediaUri.pathSegments.last()
        var dir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
        var targetFile = dir.resolve(fileName)
        if(targetFile.exists()){
            Log.w(BaseActivity.TAG,"file $targetFile already exists, omit downloading")
            return fileName
        }

        val conn = URL(params.first()).openConnection() as HttpURLConnection
        try{
            conn.connect()
            if(conn.responseCode / 100 != 2) throw IOException("STATUS CODE: ${conn.responseCode}")
            val sum = (conn.getHeaderField("Content-Length")
                    ?: conn.getHeaderField("content-length")
                    ?: "0").toInt()
            val ctype = (conn.getHeaderField("Content-Type")
                    ?: conn.getHeaderField("content-type")
                    ?: "Unknown")
            Log.i(BaseActivity.TAG,"downloading $mediaUri, Content-Type: $ctype, Content-Length: $sum")

            var current = 0
            val buff = ByteArray(1 shl  12) //4096
            var len:Int
            val buffered = conn.inputStream.buffered(1 shl 12)
            buffered.apply {

                val outBuffered = targetFile.outputStream().buffered(1 shl 12)
                outBuffered.use {
                    while(true){
                        len = this.read(buff)
                        if(len <= 0) break
                        it.write(buff,0,len)
                        current += len
                        if(sum > 0) publishProgress(current.toString(),sum.toString())
                    }
                }

            }

            return fileName
        }catch(ex:IOException){
            Log.e(BaseActivity.TAG,"fetch tweet get error: ${ex.message}",ex)
            targetFile.delete()
            return "error:${ex.message}"
        }finally{
            try{conn.inputStream.close()}catch(ex:IOException){/*ignore*/}
        }

    }
}

interface MediaDownloadListener{
    /**
     * 总字节数 与 当前下载好的字节数
     */
    fun progress(current:Int,total:Int)

    fun fail(msg:String)
    /**
     * download succeed
     */
    fun ok(fileName:String)
}