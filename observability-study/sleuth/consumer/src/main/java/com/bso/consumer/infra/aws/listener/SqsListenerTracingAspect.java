package com.bso.consumer.infra.aws.listener;

import brave.Span;
import brave.Tracer;
import brave.Tracing;
import brave.propagation.TraceContext;
import brave.propagation.TraceContextOrSamplingFlags;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.messaging.Message;
import org.springframework.messaging.MessageHeaders;
import org.springframework.stereotype.Component;

@Aspect
@Component
public class SqsListenerTracingAspect {
    private final Tracer tracer;
    private final TraceContext.Extractor<MessageHeaders> extractor;

    public SqsListenerTracingAspect(Tracer tracer, Tracing tracing) {
        this.tracer = tracer;
        extractor = tracing.propagation().extractor((request, key) -> request.get(key, String.class));
    }

    @Around("@annotation(io.awspring.cloud.messaging.listener.annotation.SqsListener)")
    public Object listenWithTrace(ProceedingJoinPoint joinPoint) throws Throwable {
        Object[] args = joinPoint.getArgs();
        if (args == null || args.length == 0) return joinPoint.proceed(args);

        Message<?> message = (Message<?>) args[0];
        TraceContextOrSamplingFlags ctx = extractor.extract(message.getHeaders());
        Span span = tracer.nextSpan(ctx);
        try (Tracer.SpanInScope ignored = tracer.withSpanInScope(span)) {
            return joinPoint.proceed(args);
        } finally {
            span.finish();
        }
    }

}
