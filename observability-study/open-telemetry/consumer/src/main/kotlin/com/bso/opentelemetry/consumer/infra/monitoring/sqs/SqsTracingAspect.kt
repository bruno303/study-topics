package com.bso.opentelemetry.consumer.infra.monitoring.sqs

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.trace.Span
import io.opentelemetry.api.trace.Tracer
import io.opentelemetry.context.Context
import org.aspectj.lang.ProceedingJoinPoint
import org.aspectj.lang.annotation.Around
import org.aspectj.lang.annotation.Aspect
import org.springframework.stereotype.Component
import software.amazon.awssdk.services.sqs.model.Message


@Aspect
@Component
class SqsTracingAspect(
    private val tracer: Tracer,
    private val openTelemetry: OpenTelemetry
) {

    @Around("@annotation(io.awspring.cloud.sqs.annotation.SqsListener)")
    fun listenWithTrace(joinPoint: ProceedingJoinPoint): Any? {
        val context: Context = openTelemetry.propagators.textMapPropagator.extract(
            Context.current(),
            (joinPoint.args[0] as Message),
            SqsTracingExtractor
        )

        val className: String = joinPoint.signature.declaringType.simpleName
        var proceed: Any? = null
        runWithTracingContext(context, className) {
            proceed = joinPoint.proceed()
        }
        return proceed
    }

    private fun runWithTracingContext(
        context: Context,
        className: String,
        block: () -> Unit
    ) {
        context.makeCurrent().use {
            val span = buildSpan(context, className)
            span.makeCurrent().use {
                try {
                    block.invoke()
                } finally {
                    span.end()
                }
            }
        }
    }

    private fun buildSpan(context: Context, className: String): Span =
        tracer.spanBuilder("sqs.listener.$className")
            .startSpan()
            .also { it.storeInContext(context) }
}