package com.bso.passwordhash

import java.math.BigInteger
import java.security.MessageDigest

class PasswordHasher(
    private val saltGenerator: SaltGenerator
) {

//    suspend fun encrypt(text: String): String {
//        return runAsync {
//            BCrypt.withDefaults().hash(
//                COST,
//                SALT,
//                text.toByteArray(Charsets.UTF_8)
//            ).toString()
//        }
//    }

        fun encrypt(text: String, salt: ByteArray? = null): String = text.sha256(salt)

        fun validate(hash: String, plainText: String): Boolean {
            val salt = hash.split(":")[0].toByteArray()
            return encrypt(plainText, salt) == hash
        }

//    fun String.md5(): String {
//        val md = MessageDigest.getInstance("MD5")
//        return BigInteger(1, md.digest(toByteArray())).toString(16).padStart(32, '0')
//    }
//
//    fun String.sha1(): String {
//        val md = MessageDigest.getInstance("SHA-1")
//        return BigInteger(1, md.digest(toByteArray())).toString(16).padStart(32, '0')
//    }

        private fun String.sha256(salt: ByteArray? = null): String {
            val generatedSalt = salt ?: saltGenerator.generate()
            val md = MessageDigest.getInstance("SHA-256")
            md.update(generatedSalt)
            val hash = md.digest(this.toByteArray()).toUtf8String().toHex()
            return "${generatedSalt.toUtf8String()}:$hash"
        }

        private fun String.toHex(): String =
            BigInteger(1, this.toByteArray()).toString(16)

        private fun ByteArray.toUtf8String() = toString(Charsets.UTF_8)

//        @OptIn(DelicateCoroutinesApi::class)
//        private suspend fun <T> runAsync(block: () -> T): T = GlobalScope
//            .async { block() }
//            .await()

        companion object {
            private const val COST: Int = 4
            private val SALT: ByteArray = "[259ca4".toByteArray(Charsets.UTF_16)
        }
}