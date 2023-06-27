package com.bso.service1.externalcommunication.service2.auth

import feign.RequestInterceptor
import feign.RequestTemplate
import org.springframework.security.core.context.SecurityContextHolder
import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationToken

class Service2AuthPropagateToken : RequestInterceptor {
    override fun apply(template: RequestTemplate) {
        val token = (SecurityContextHolder.getContext().authentication as JwtAuthenticationToken).token
        template.header("Authorization", "Bearer ${token.tokenValue}")
    }
}