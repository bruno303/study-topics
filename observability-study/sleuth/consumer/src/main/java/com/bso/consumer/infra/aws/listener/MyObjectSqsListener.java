package com.bso.consumer.infra.aws.listener;

import com.bso.consumer.infra.aws.sender.MessageSender;
import com.bso.consumer.model.MyObject;
import io.awspring.cloud.messaging.listener.annotation.SqsListener;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.messaging.Message;
import org.springframework.stereotype.Component;

@Component
@Slf4j
@RequiredArgsConstructor
public class MyObjectSqsListener {
    private final MessageSender messageSender;

    @Value("${app.queues.my-queue-response}")
    private String queueResponse;

    @SqsListener("${app.queues.my-queue}")
    public void listen(Message<MyObject> message) {
        log.info("Message received : {}", message.getPayload());
        messageSender.send(message.getPayload(), queueResponse);
    }
}
