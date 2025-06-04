package com.bso.postgressharding.infra.repository.impl

import com.bso.postgressharding.infra.entities.User
import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository

interface UserRepositoryJpa : JpaRepository<User, String>