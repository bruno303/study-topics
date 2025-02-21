package com.bso.k8stest.application.broadcast

import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.boot.web.client.RestTemplateBuilder
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpMethod
import org.springframework.stereotype.Service
import org.springframework.util.MultiValueMap
import org.springframework.web.client.RestTemplate

@Service
class BroadcastService {
    private val logger: Logger = LoggerFactory.getLogger(this::class.java)

    fun execute(
        command: BroadcastCommand,
        originalRequestHeaders: Map<String, String>
    ) {
        val restTemplate: RestTemplate = RestTemplateBuilder().build()

        command.urls.forEach { uri ->
            logger.info("Sending request to <[{}] {}>", command.method, uri)
            val response: String? = when (HttpMethod.valueOf(command.method)) {
                HttpMethod.POST -> {
                    val httpHeaders: MultiValueMap<String, String> = HttpHeaders()
                        .apply { fulfillHeaders(command, originalRequestHeaders, this) }
                    val entity: HttpEntity<String>? = command.body?.let { HttpEntity(it, httpHeaders) }
                    restTemplate.postForObject(uri, entity, String::class.java)
                }
                HttpMethod.GET -> restTemplate.getForObject(uri, String::class.java)
                else -> throw IllegalArgumentException("Method ${command.method} not expected")
            }
            logger.info("Response for uri <{}>: <{}>", uri, response)
        }
    }

    private fun fulfillHeaders(
        command: BroadcastCommand,
        headers: Map<String, String>,
        httpHeaders: HttpHeaders
    ) {
        command.addHeaders?.let {
            logger.debug("Add Headers were informed: <{}>", it)
            it.forEach { h ->
                logger.trace("Adding header: <{}:{}>", h.key, h.value)
                httpHeaders.add(h.key, h.value)
            }
        }
        command.bypassHeaders?.let {
            logger.debug("ByPass Headers were informed: <{}>", it)
            it.forEach { h ->
                val currentRequestHeaderValue = headers[h]
                logger.trace("ByPass header is <{}>. Current request value is <{}>", h, currentRequestHeaderValue)
                currentRequestHeaderValue?.let { value -> httpHeaders.add(h, value)  }
            }
        }
    }
}
