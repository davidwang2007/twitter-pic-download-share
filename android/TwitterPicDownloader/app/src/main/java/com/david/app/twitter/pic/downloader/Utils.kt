package com.david.app.twitter.pic.downloader

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.support.v4.app.ActivityCompat
import android.util.Log
import android.webkit.MimeTypeMap
import java.io.File


class Utils{
    companion object {
        val settings:String = "settings"
        /**
         * pref
         */
        fun prefEdit(ctx:Context) = ctx.getSharedPreferences(settings,Context.MODE_PRIVATE).edit()
        fun pref(ctx:Context) = ctx.getSharedPreferences(settings,Context.MODE_PRIVATE)
        fun verifyStoragePermissions(ctx: Activity){
            val permission = ActivityCompat.checkSelfPermission(ctx, Manifest.permission.WRITE_EXTERNAL_STORAGE)
            if (permission != PackageManager.PERMISSION_GRANTED){
                ActivityCompat.requestPermissions(ctx,
                        arrayOf(Manifest.permission.READ_EXTERNAL_STORAGE,Manifest.permission.WRITE_EXTERNAL_STORAGE),
                        1)
            }
        }

        /**
         * 直接弹出分享图片
         */
        fun sendMedia(ctx:Context,file:File){
            val map = MimeTypeMap.getSingleton()
            val ext = MimeTypeMap.getFileExtensionFromUrl(file.name)
            var type = map.getMimeTypeFromExtension(ext)
            if(type == null) type = "*/*"
            //val path = MediaStore.Images.Media.insertImage(ctx.contentResolver,file.absolutePath,file.name,"download from twitter")
            Log.i(BaseActivity.TAG,"try to share $file, type: $type")
            val data = Uri.fromFile(file)
            val intent = Intent(Intent.ACTION_SEND)
            intent.setType(type)
            intent.putExtra(Intent.EXTRA_STREAM,data)
            intent.putExtra(Intent.EXTRA_SUBJECT,"Pic From Twitter")
            intent.putExtra(Intent.EXTRA_TEXT,"Share Pic From Twitter")
            ctx.startActivity(Intent.createChooser(intent,"SELECT"))
        }

    }
}

