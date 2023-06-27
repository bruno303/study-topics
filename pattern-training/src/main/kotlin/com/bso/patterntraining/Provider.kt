package com.bso.patterntraining

import com.bso.patterntraining.templatemedthod.Payload

interface Provider {
    fun execute(payload: Payload)
}