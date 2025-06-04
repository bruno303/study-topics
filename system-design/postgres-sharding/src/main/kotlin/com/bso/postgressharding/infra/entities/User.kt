package com.bso.postgressharding.infra.entities

import jakarta.persistence.Entity
import jakarta.persistence.Id

@Entity
data class User(
    @Id
    val id: String,
    val name: String,
    val email: String
)
