package com.bso.postgressharding.infra.config

import org.springframework.boot.jdbc.DataSourceBuilder
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import javax.sql.DataSource

@Configuration
class ShardDataSourceConfig(
    private val properties: ShardDatabaseProperties
) {

    @Bean
    fun shardDataSources(): Map<String, DataSource> {
        return properties.all().mapValues { (_, props) ->
            DataSourceBuilder.create()
                .url(props.url)
                .username(props.username)
                .password(props.password)
                .driverClassName(props.driverClassName ?: "org.postgresql.Driver")
                .build()
        }
    }
}
