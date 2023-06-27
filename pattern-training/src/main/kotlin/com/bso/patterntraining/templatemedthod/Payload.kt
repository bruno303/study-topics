package com.bso.patterntraining.templatemedthod

import kotlin.random.Random

data class Payload(
    val name: String,
    val number: Int,
    val boolean: Boolean,
    val type: Type
) {
    companion object {
        fun random() = Payload(
            name = "name",
            number = Random.nextInt(),
            boolean = Random.nextBoolean(),
            type = Type.values().toList().random()
        )
    }
}

enum class Type {
    TYPE_A,
    TYPE_B
}