package com.bso.opentelemetry.consumer.infra.monitoring

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.common.Attributes
import io.opentelemetry.api.trace.Tracer
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator
import io.opentelemetry.context.propagation.ContextPropagators
import io.opentelemetry.exporter.zipkin.ZipkinSpanExporter
import io.opentelemetry.sdk.OpenTelemetrySdk
import io.opentelemetry.sdk.resources.Resource
import io.opentelemetry.sdk.trace.SdkTracerProvider
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class TracingConfiguration {
    private val appName: String = "com.bso.opentelemetry.consumer.ConsumerApplication"

    @Bean
    fun openTelemetry(): OpenTelemetry {
        val resource: Resource = Resource.getDefault()
            .merge(Resource.create(Attributes.of(ResourceAttributes.SERVICE_NAME, appName)))

        val sdkTracerProvider: SdkTracerProvider = ZipkinSpanExporter
            .builder()
            .setEndpoint("http://localhost:9411/api/v2/spans")
            .build().let {
                SdkTracerProvider.builder()
                    .addSpanProcessor(BatchSpanProcessor.builder(it).build())
                    .setResource(resource)
                    .build()
            }

        Runtime.getRuntime().addShutdownHook(Thread(sdkTracerProvider::close))

        return OpenTelemetrySdk.builder()
            .setTracerProvider(sdkTracerProvider)
            .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
            .buildAndRegisterGlobal()
    }

    @Bean
    fun tracer(openTelemetry: OpenTelemetry): Tracer =
        openTelemetry.getTracer(appName)
}