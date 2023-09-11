package com.bso.k8stest.application.broadcast

data class BroadcastCommand(
    val urls: List<String>,
    val method: String,
    val body: String?,
    val addHeaders: Map<String, String>?,
    val bypassHeaders: List<String>?
)
