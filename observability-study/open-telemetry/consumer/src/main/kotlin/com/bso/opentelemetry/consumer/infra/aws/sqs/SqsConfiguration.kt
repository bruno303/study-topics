package com.bso.opentelemetry.consumer.infra.aws.sqs

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.Primary
import org.springframework.context.annotation.Profile
import software.amazon.awssdk.services.sqs.SqsAsyncClient
import software.amazon.awssdk.services.sqs.SqsClient
import java.net.URI

@Configuration
class SqsConfiguration {
    @Bean
    @Profile("local")
    @Primary
    fun sqsAsyncClient(): SqsAsyncClient {
        return SqsAsyncClient
            .builder()
            .endpointOverride(URI.create("http://localhost:4566"))
            .build()
    }

    @Bean
    @Profile("local")
    @Primary
    fun sqsClient(): SqsClient {
        return SqsClient
            .builder()
            .endpointOverride(URI.create("http://localhost:4566"))
            .build()
    }
}