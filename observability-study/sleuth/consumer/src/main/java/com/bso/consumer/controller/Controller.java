package com.bso.consumer.controller;

import com.bso.consumer.infra.aws.sender.MessageSender;
import com.bso.consumer.model.MyObject;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/object")
@RequiredArgsConstructor
@Slf4j
public class Controller {

    private final MessageSender messageSender;

    @Value("${app.queues.my-queue-response}")
    private String queue;

    @PostMapping
    public void request(@RequestBody MyObject myObject) {
        log.info("[{}] Received", myObject.id());
        messageSender.send(myObject, queue);
    }
}
