package com.bso.patterntraining.templatemedthod

import org.springframework.stereotype.Component

@Component
class ExecutorService {
    private val providerFactory = ProviderFactory()

    fun execute(payload: Payload) {
        providerFactory
            .create(type = payload.type)
            .execute(payload = payload)
    }
}