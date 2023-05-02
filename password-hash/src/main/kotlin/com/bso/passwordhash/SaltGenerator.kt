package com.bso.passwordhash

import java.security.SecureRandom

class SaltGenerator {
    private val secureRandom = SecureRandom()

    fun generate(): ByteArray {
        return ByteArray(16).also {
            secureRandom.nextBytes(it)
        }
    }
}