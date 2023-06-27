package com.bso.service1.externalcommunication.service2

import org.springframework.stereotype.Service

@Service
class Service2Service(
    private val service2Feign: Service2Feign
) {

    fun getMessage(): String {
        return service2Feign.protected().message
    }
}