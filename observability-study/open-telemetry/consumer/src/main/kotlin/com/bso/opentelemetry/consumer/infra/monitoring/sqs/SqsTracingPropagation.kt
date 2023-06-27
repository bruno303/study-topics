package com.bso.opentelemetry.consumer.infra.monitoring.sqs

import io.opentelemetry.context.propagation.TextMapGetter
import io.opentelemetry.context.propagation.TextMapSetter
import software.amazon.awssdk.services.sqs.model.Message
import software.amazon.awssdk.services.sqs.model.MessageAttributeValue
import software.amazon.awssdk.services.sqs.model.SendMessageRequest

object SqsTracingExtractor : TextMapGetter<Message> {
    override fun keys(carrier: Message): MutableIterable<String> =
        carrier.messageAttributes().keys

    override fun get(carrier: Message?, key: String): String? =
        carrier?.messageAttributes()?.get(key)?.stringValue()
}

object SqsTracingInjector: TextMapSetter<MutableMap<String, MessageAttributeValue>> {
    override fun set(carrier: MutableMap<String, MessageAttributeValue>?, key: String, value: String) {
        val attribute = MessageAttributeValue.builder().stringValue(value).dataType("String").build()
        carrier?.put(key, attribute)
    }
}