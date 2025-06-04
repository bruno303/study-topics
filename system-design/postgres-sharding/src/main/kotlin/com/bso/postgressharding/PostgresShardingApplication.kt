package com.bso.postgressharding

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration
import org.springframework.boot.runApplication

@SpringBootApplication(exclude = [DataSourceAutoConfiguration::class])
class PostgresShardingApplication

fun main(args: Array<String>) {
	runApplication<PostgresShardingApplication>(*args)
}
