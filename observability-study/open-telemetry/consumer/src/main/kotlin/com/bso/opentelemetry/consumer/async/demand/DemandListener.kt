package com.bso.opentelemetry.consumer.async.demand

import com.bso.opentelemetry.consumer.domain.entity.Demand
import com.bso.opentelemetry.consumer.infra.aws.sqs.SqsMessageSender
import com.fasterxml.jackson.databind.ObjectMapper
import io.awspring.cloud.sqs.annotation.SqsListener
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Component
import software.amazon.awssdk.services.sqs.model.Message
import java.util.*

@Component
class DemandListener(
    private val objectMapper: ObjectMapper,
    private val sqsMessageSender: SqsMessageSender,
    @Value("\${aws.sqs.queues.demand-response}")
    private val responseQueue: String
) {
    private val logger: Logger by lazy { LoggerFactory.getLogger(this::class.java) }

    @SqsListener("\${aws.sqs.queues.demand}")
    fun listen(message: Message) {
        val demandDto = objectMapper.readValue(message.body(), DemandDto::class.java)
        val demand = demandDto.toDemand()
        logger.info("Received message for process demand: {}", demand)
        sqsMessageSender.send(responseQueue, demand)
    }
}

private fun DemandDto.toDemand(): Demand = Demand(
    id = UUID.randomUUID(),
    name = name,
    age = age
)