package com.bso.passwordhash

import io.mockk.every
import io.mockk.mockk
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class PasswordHasherTest {
    private val saltGeneratorMock: SaltGenerator = mockk()
    private val subject = PasswordHasher(saltGeneratorMock)

    @BeforeEach
    fun setup() {
        every { saltGeneratorMock.generate() } returns SALT.toByteArray() // [54, 53, 50, 51, 52, 49]
    }

    @Test
    fun `should generate a hash for password`() {
        val encryptedText: String = subject.encrypt(PLAIN_PASSWORD)
        assertEquals(HASH, encryptedText)
    }

    @Test
    fun `should validate password correctly`() {
        assertTrue(subject.validate(HASH, PLAIN_PASSWORD))
    }

    @Test
    fun `should return false when plainPassword does not match`() {
        val invalidPassword = "aBc1234." // correct password "aBc1234;"
        assertFalse(subject.validate(HASH, invalidPassword))
    }

    companion object {
        private const val SALT = "652341"
        private const val HASH = "$SALT:1e3a490eefbfbdefbfbdefbfbdefbfbdefbfbd64275eefbfbde48b97efbfbd49efbfbd1eefbfbdefbfbdefbfbdefbfbd53411cefbfbd01efbfbd295e"
        private const val PLAIN_PASSWORD = "aBc1234;"
    }
}