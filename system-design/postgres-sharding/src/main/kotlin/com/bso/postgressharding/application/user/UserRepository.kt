package com.bso.postgressharding.application.user

import com.bso.postgressharding.domain.entities.User
import java.util.UUID

interface UserRepository {
    fun save(user: User): User
    fun findById(id: UUID): User?
}
