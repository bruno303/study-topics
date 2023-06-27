package com.bso.consumer.infra.aws.sender;

public interface MessageSender {
    <T> void send(T message, String queue);
}
