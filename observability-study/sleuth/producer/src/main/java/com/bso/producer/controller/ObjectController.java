package com.bso.producer.controller;

import com.bso.producer.infra.aws.sender.SqsMessageSender;
import com.bso.producer.infra.feign.Service2Feign;
import com.bso.producer.model.MyObject;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RequiredArgsConstructor
@RestController
@RequestMapping("/object")
@Slf4j
public class ObjectController {

    private final Service2Feign service2;
    private final SqsMessageSender messageSender;

    @Value("${app.queues.my-queue}")
    private String queue;

    @PostMapping
    public void request(@RequestBody MyObject myObject) {
        log.info("[{}] Requesting", myObject.id());
        service2.sendToService2(myObject);
    }

    @PostMapping("async")
    public void requestAsync(@RequestBody MyObject myObject) {
        log.info("[{}] Requesting", myObject.id());
        messageSender.send(myObject, queue);
    }
}
