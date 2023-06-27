package com.bso.patterntraining.templatemedthod

import com.bso.patterntraining.Provider
import org.slf4j.Logger
import org.slf4j.LoggerFactory

abstract class BaseProvider : Provider {
    private val logger: Logger by lazy { LoggerFactory.getLogger(BaseProvider::class.java) }

    override fun execute(payload: Payload) {
        logger.info("Starting execution")
        runCatching { doExecute(payload) }
            .onFailure { logger.error("Error", it) }
            .onSuccess { logger.info("Success") }
    }

    abstract fun doExecute(payload: Payload)
}