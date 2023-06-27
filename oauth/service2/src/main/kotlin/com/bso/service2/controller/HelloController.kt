package com.bso.service2.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RequestMapping("hello")
@RestController
class HelloController {

    @GetMapping("/unprotected")
    fun unprotectedHello(): Response = Response()

    @GetMapping("/protected")
    fun protectedHello(): Response = Response()

    @GetMapping("/denied")
    fun deniedHello(): Response = Response()
}

data class Response(val message: String = "Hello world from service 2")