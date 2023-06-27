package com.bso.producer.infra.aws.listener;

import com.bso.producer.model.MyObject;
import io.awspring.cloud.messaging.listener.annotation.SqsListener;
import lombok.extern.slf4j.Slf4j;
import org.springframework.messaging.Message;
import org.springframework.stereotype.Component;

@Component
@Slf4j
public class MyObjectSqsListener {
    @SqsListener("${app.queues.my-queue-response}")
    public void listen(Message<MyObject> message) {
        log.info("Message received: {}", message.getPayload());
    }
}
