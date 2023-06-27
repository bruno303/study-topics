package com.bso.service1.externalcommunication.service2.auth

import com.fasterxml.jackson.annotation.JsonProperty

data class TokenResponse(
    @JsonProperty("access_token")
    val token: String
)