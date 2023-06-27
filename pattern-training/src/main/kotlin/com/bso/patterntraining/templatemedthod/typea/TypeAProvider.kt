package com.bso.patterntraining.templatemedthod.typea

import com.bso.patterntraining.templatemedthod.BaseProvider
import com.bso.patterntraining.templatemedthod.Payload
import org.slf4j.Logger
import org.slf4j.LoggerFactory

class TypeAProvider : BaseProvider() {
    private val logger: Logger by lazy { LoggerFactory.getLogger(BaseProvider::class.java) }

    override fun doExecute(payload: Payload) {
        logger.info("TypeAProvider")
    }
}