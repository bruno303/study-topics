package com.bso.postgressharding.web.controllers

import com.bso.postgressharding.application.user.UserService
import com.bso.postgressharding.domain.entities.User
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import java.util.UUID

@RestController
@RequestMapping("/users")
class UserController(private val userService: UserService) {

    @PostMapping
    fun create(@RequestBody user: UserCreationRequest) = userService.save(
        User(
            UUID.randomUUID(),
            user.name,
            user.email
        )
    )

    @GetMapping("/{id}")
    fun read(@PathVariable id: String) = userService.findById(UUID.fromString(id))
}

data class UserCreationRequest(
    val name: String,
    val email: String
)