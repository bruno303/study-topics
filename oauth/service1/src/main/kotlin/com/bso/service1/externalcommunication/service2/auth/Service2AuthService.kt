package com.bso.service1.externalcommunication.service2.auth

import org.springframework.stereotype.Service

@Service
class Service2AuthService(
    private val service2AuthFeign: Service2AuthFeign
) {
    fun login(): String {
        val tokenResponse: TokenResponse = service2AuthFeign.login(
            authorization = "Basic aW50ZXJuYWxfY2xpZW50OjA1NjlkNTY1LTEzNzktNGNjYy1hNTk4LThlODg1ZTA5NmE1ZQ==",
            body = mapOf(
                "grant_type" to "client_credentials",
                "scope" to "teste"
            )
        )
        return tokenResponse.token
    }
}