package com.bso.opentelemetry.consumer.infra.aws.sqs

import com.bso.opentelemetry.consumer.infra.monitoring.sqs.SqsTracingInjector
import com.fasterxml.jackson.databind.ObjectMapper
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.context.Context
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Component
import software.amazon.awssdk.services.sqs.SqsAsyncClient
import software.amazon.awssdk.services.sqs.model.GetQueueUrlRequest
import software.amazon.awssdk.services.sqs.model.MessageAttributeValue
import software.amazon.awssdk.services.sqs.model.SendMessageRequest

@Component
class SqsMessageSender(
    private val sqsAsyncClient: SqsAsyncClient,
    private val objectMapper: ObjectMapper,
    private val openTelemetry: OpenTelemetry
) {
    private val logger: Logger by lazy { LoggerFactory.getLogger(this::class.java) }
    private val cacheLock = Any()
    private val cache: Cache<String, String> = Cache()

    fun send(queueName: String, message: Any) {
        val queueUrl: String = getQueueFromCacheOrSdk(queueName)
        val request = SendMessageRequest.builder()
            .queueUrl(queueUrl)
            .messageBody(objectMapper.writeValueAsString(message))
            .messageAttributes(buildMessageAttributes())
            .build()
        sqsAsyncClient.sendMessage(request)
    }

    private fun getQueueFromCacheOrSdk(queueName: String): String {
        if (cache.contains(queueName)) {
            logger.debug("Using queueUrl from cache. QueueName: {}", queueName)
            return cache.get(queueName)!!
        }
        return synchronized(cacheLock) {
            if (!cache.contains(queueName)) { // double check to avoid dirty read
                logger.debug("Seaching queueUrl with Amazon SDK. QueueName: {}", queueName)
                cache.add(queueName, searchQueueUrl(queueName))
            }
            cache.get(queueName)!!
        }
    }

    // Do sync because will be done only one time per queue and instance running
    private fun searchQueueUrl(queueName: String): String {
        return sqsAsyncClient.getQueueUrl(GetQueueUrlRequest.builder()
            .queueName(queueName)
            .build()
        ).join().queueUrl()
    }

    private fun buildMessageAttributes(): Map<String, MessageAttributeValue> {
        val attributes: MutableMap<String, MessageAttributeValue> = mutableMapOf()

        openTelemetry.propagators.textMapPropagator.inject(
            Context.current(),
            attributes,
            SqsTracingInjector
        )

        return attributes.toMap()
    }
}

class Cache<K, V> {
    private val cache: MutableMap<K, V> = mutableMapOf()
    fun contains(key: K): Boolean = cache.containsKey(key)
    fun get(key: K): V? = cache[key]
    fun add(key: K, value: V) { cache[key] = value }
}