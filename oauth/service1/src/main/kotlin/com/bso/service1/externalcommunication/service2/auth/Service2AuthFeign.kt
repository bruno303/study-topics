package com.bso.service1.externalcommunication.service2.auth

import org.springframework.cloud.openfeign.FeignClient
import org.springframework.http.MediaType
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestHeader

@FeignClient(
    name = "service2AuthFeign",
    url = "http://localhost:8081/auth/realms/test/protocol/openid-connect/token"
)
interface Service2AuthFeign {
    @PostMapping(consumes = [MediaType.APPLICATION_FORM_URLENCODED_VALUE])
    fun login(
        @RequestHeader("Authorization") authorization: String,
        body: Map<String, Any>
    ): TokenResponse
}