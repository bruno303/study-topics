package com.bso.k8stest.web

import com.bso.k8stest.application.broadcast.BroadcastCommand
import com.bso.k8stest.application.broadcast.BroadcastService
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.web.bind.annotation.*

@RestController
@RequestMapping("/")
class HelloController(
    private val broadcastService: BroadcastService
) {
    private val logger: Logger = LoggerFactory.getLogger(this::class.java)

    @GetMapping
    fun hello(@RequestHeader headers: Map<String, String>): Message {
        logger.debug("Received request for 'hello' with headers <{}>", headers)
        logger.trace("This is a TRACE level message");
        logger.debug("This is a DEBUG level message");
        logger.info("This is an INFO level message");
        logger.warn("This is a WARN level message");
        logger.error("This is an ERROR level message");
        return Message("Hello World")
    }

    @PostMapping("/broadcast")
    fun broadcast(
        @RequestBody request: BroadcastCommand,
        @RequestHeader headers: Map<String, String>
    ): Message {
        logger.debug("Received request for broadcast with headers: <{}> and body <{}>", headers, request)
        broadcastService.execute(command = request, headers)
        return Message("Broadcast done successfully")
    }
}

data class Message(val message: String)
