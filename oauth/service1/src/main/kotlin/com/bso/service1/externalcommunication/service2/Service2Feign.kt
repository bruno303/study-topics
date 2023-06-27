package com.bso.service1.externalcommunication.service2

import com.bso.service1.controller.Response
import com.bso.service1.externalcommunication.service2.auth.Service2AuthConfiguration
import org.springframework.cloud.openfeign.FeignClient
import org.springframework.web.bind.annotation.GetMapping

@FeignClient(
    name = "service2Feign",
    url = "http://localhost:8082/hello",
    configuration = [Service2AuthConfiguration::class]
)
interface Service2Feign {

    @GetMapping("protected")
    fun protected(): Response
}