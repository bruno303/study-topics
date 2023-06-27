package com.bso.service1.controller

import com.bso.service1.externalcommunication.service2.Service2Service
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RequestMapping("hello")
@RestController
class HelloController(
    private val service2Service: Service2Service
) {

    @GetMapping("/unprotected")
    fun unprotectedHello(): Response = Response()

    @GetMapping("/protected")
    fun protectedHello(): Response = Response(service2Service.getMessage())

    @GetMapping("/denied")
    fun deniedHello(): Response = Response()
}

data class Response(val message: String = "Hello World")