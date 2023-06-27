package com.bso.service1.config

import feign.codec.Encoder
import feign.form.FormEncoder
import org.springframework.beans.factory.ObjectFactory
import org.springframework.beans.factory.config.BeanDefinition.SCOPE_PROTOTYPE
import org.springframework.boot.autoconfigure.http.HttpMessageConverters
import org.springframework.cloud.openfeign.support.SpringEncoder
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.Primary
import org.springframework.context.annotation.Scope

@Configuration
class FeignConfiguration(
    private val messageConverters: ObjectFactory<HttpMessageConverters>
) {
    @Bean
    @Primary
    @Scope(SCOPE_PROTOTYPE)
    fun feignFormEncoder(): Encoder =
        FormEncoder(SpringEncoder(this.messageConverters))
}