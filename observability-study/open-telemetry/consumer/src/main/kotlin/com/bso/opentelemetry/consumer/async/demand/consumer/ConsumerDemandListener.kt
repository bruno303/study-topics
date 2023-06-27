package com.bso.opentelemetry.consumer.async.demand.consumer

import com.bso.opentelemetry.consumer.domain.entity.Demand
import com.fasterxml.jackson.databind.ObjectMapper
import io.awspring.cloud.sqs.annotation.SqsListener
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.stereotype.Component
import software.amazon.awssdk.services.sqs.model.Message

@Component
class ConsumerDemandListener(
    private val objectMapper: ObjectMapper
) {
    private val logger: Logger by lazy { LoggerFactory.getLogger(this::class.java) }

    /**
     * Just to simulate a consumer. It's in same project, but should be isolated as another microservice.
     */
    @SqsListener("\${aws.sqs.queues.demand-response}")
    fun listen(message: Message) {
        val demand = objectMapper.readValue(message.body(), Demand::class.java)
        logger.info("Consuming message for demand: {}", demand)
    }
}