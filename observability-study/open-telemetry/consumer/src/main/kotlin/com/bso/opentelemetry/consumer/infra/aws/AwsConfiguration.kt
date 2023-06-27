package com.bso.opentelemetry.consumer.infra.aws

import io.awspring.cloud.core.region.StaticRegionProvider
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import software.amazon.awssdk.regions.providers.AwsRegionProvider

@Configuration
class AwsConfiguration {

    @Bean
    fun regionProvider(): AwsRegionProvider = StaticRegionProvider("us-east-1")
}