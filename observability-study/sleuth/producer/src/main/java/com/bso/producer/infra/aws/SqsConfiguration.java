package com.bso.producer.infra.aws;

import com.amazonaws.auth.AWSCredentialsProvider;
import com.amazonaws.auth.DefaultAWSCredentialsProviderChain;
import com.amazonaws.client.builder.AwsClientBuilder;
import com.amazonaws.regions.DefaultAwsRegionProviderChain;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.sqs.AmazonSQSAsync;
import com.amazonaws.services.sqs.AmazonSQSAsyncClientBuilder;
import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.messaging.converter.MappingJackson2MessageConverter;
import org.springframework.messaging.converter.MessageConverter;

@Configuration
@RequiredArgsConstructor
public class SqsConfiguration {
    private final TracingRequestHandler tracingRequestHandler;

    @Bean
    @Primary
    public AmazonSQSAsync amazonSQSAsync(Regions region) {
        return AmazonSQSAsyncClientBuilder.standard()
                .withCredentials(credentialsProvider())
                .withEndpointConfiguration(new AwsClientBuilder.EndpointConfiguration("http://localhost:4566", region.getName()))
                .withRequestHandlers(tracingRequestHandler)
                .build();
    }

    @Bean
    @Primary
    public AWSCredentialsProvider credentialsProvider() {
        return DefaultAWSCredentialsProviderChain.getInstance();
    }

    @Bean
    public Regions region() {
        return Regions.fromName(new DefaultAwsRegionProviderChain().getRegion());
    }

    @Bean
    @Primary
    public MessageConverter messageConverter() {
        return new MappingJackson2MessageConverter();
    }
}
