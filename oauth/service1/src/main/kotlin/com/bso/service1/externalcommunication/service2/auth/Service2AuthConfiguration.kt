package com.bso.service1.externalcommunication.service2.auth

import feign.RequestInterceptor
import feign.RequestTemplate

class Service2AuthConfiguration(
    private val service2AuthService: Service2AuthService
) : RequestInterceptor {
    override fun apply(template: RequestTemplate) {
        val token = service2AuthService.login()
        template.header("Authorization", "Bearer $token")
    }
}