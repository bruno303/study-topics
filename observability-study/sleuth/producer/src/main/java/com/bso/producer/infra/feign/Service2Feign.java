package com.bso.producer.infra.feign;

import com.bso.producer.model.MyObject;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

@FeignClient(
        name = "service2",
        url = "${external-services.service2.url}/object"
)
public interface Service2Feign {
    @PostMapping
    void sendToService2(@RequestBody MyObject body);
}
