package com.bso.postgressharding.infra.config

import org.springframework.boot.autoconfigure.jdbc.DataSourceProperties
import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.stereotype.Component

@ConfigurationProperties(prefix = "spring.shards.datasources")
@Component
class ShardDatabaseProperties {
    var shard1: DataSourceProperties = DataSourceProperties()
    var shard2: DataSourceProperties = DataSourceProperties()
    var shard3: DataSourceProperties = DataSourceProperties()

    fun all(): Map<String, DataSourceProperties> = mapOf(
        "shard1" to shard1,
        "shard2" to shard2,
        "shard3" to shard3,
    )
}

@ConfigurationProperties(prefix = "spring.jpa.hibernate")
@Component
class SpringJpaProperties {
    var ddlAuto: String = ""
    var dialect: String = ""
}