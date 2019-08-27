package com.david.app.twitter.pic.downloader

import android.support.v7.app.AppCompatActivity
import com.google.android.gms.ads.AdRequest
import com.google.android.gms.ads.AdView
import com.google.android.gms.ads.MobileAds

abstract class BaseActivity:AppCompatActivity(){

    protected fun initAd(adView: AdView){
        MobileAds.initialize(this,getString(R.string.admob_app_id))
        // Load an ad into the AdMob banner view.
        val adRequest = AdRequest.Builder()
                //.setRequestAgent("android_studio:ad_template")
                .build()
        adView.loadAd(adRequest)

    }

    companion object {
        val TAG = "TPD"
    }
}