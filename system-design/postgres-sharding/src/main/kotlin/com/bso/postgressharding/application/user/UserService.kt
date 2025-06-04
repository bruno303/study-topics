package com.bso.postgressharding.application.user

import com.bso.postgressharding.domain.entities.User
import org.springframework.stereotype.Service
import java.util.UUID

@Service
class UserService(
    private val userRepository: UserRepository
) {
    fun save(user: User): User = userRepository.save(user)
    fun findById(userId: UUID): User? = userRepository.findById(userId)
}
