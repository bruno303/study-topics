package com.bso.patterntraining.templatemedthod

import com.bso.patterntraining.Provider
import com.bso.patterntraining.templatemedthod.typea.TypeAProvider

class ProviderFactory {
    fun create(type: Type): Provider =
        when(type) {
            Type.TYPE_A -> TypeAProvider()
            else -> object : Provider {
                override fun execute(payload: Payload) {
                    TODO("Not yet implemented")
                }
            }
        }
}