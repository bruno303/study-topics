package com.bso.opentelemetry.consumer.domain.entity

import java.util.UUID

data class Demand(
    val id: UUID? = null,
    val name: String,
    val age: Int
)