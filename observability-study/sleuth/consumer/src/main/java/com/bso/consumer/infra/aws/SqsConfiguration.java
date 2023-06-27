package com.bso.consumer.infra.aws;

import com.amazonaws.auth.AWSCredentialsProvider;
import com.amazonaws.auth.DefaultAWSCredentialsProviderChain;
import com.amazonaws.client.builder.AwsClientBuilder;
import com.amazonaws.regions.DefaultAwsRegionProviderChain;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.sqs.AmazonSQSAsync;
import com.amazonaws.services.sqs.AmazonSQSAsyncClientBuilder;
import io.awspring.cloud.messaging.core.QueueMessagingTemplate;
import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.messaging.converter.MappingJackson2MessageConverter;
import org.springframework.messaging.converter.MessageConverter;
import org.springframework.stereotype.Component;

@Configuration
@Component
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
    public QueueMessagingTemplate queueMessagingTemplate(AmazonSQSAsync amazonSQSAsync) {
        return new QueueMessagingTemplate(amazonSQSAsync);
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
