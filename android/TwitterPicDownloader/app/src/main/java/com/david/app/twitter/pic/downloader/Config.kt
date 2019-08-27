package com.david.app.twitter.pic.downloader

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import java.io.ByteArrayOutputStream
import java.io.InputStream

inline fun <reified T> Gson.fromJson(src: InputStream) =
    ByteArrayOutputStream().let {
        src.copyTo(it)
        this.fromJson<T>(it.toString("UTF-8"),object:TypeToken<T>(){}.type)
    }

