package com.bso.consumer.infra.aws.sender;

import brave.Tracer;
import brave.Tracing;
import brave.propagation.Propagation;
import brave.propagation.TraceContext;
import io.awspring.cloud.messaging.core.QueueMessagingTemplate;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Component
public class SqsMessageSender implements MessageSender {
    private final QueueMessagingTemplate queueMessagingTemplate;
    private final Tracer tracer;
    private final TraceContext.Injector<Map<String, Object>> injector;

    public SqsMessageSender(Tracing tracing, QueueMessagingTemplate queueMessagingTemplate, Tracer tracer) {
        this.queueMessagingTemplate = queueMessagingTemplate;
        this.tracer = tracer;
        Propagation.Setter<Map<String, Object>, String> setter = (request, key, value) -> request.put(key, value);
        this.injector = tracing.propagation().injector(setter);
    }

    @Override
    public <T> void send(T message, String queue) {
        Map<String, Object> headers = new HashMap<>();
        injector.inject(tracer.currentSpan().context(), headers);
        queueMessagingTemplate.convertAndSend(queue, message, headers);
    }
}
