package com.bso.postgressharding.application.migration

import org.flywaydb.core.Flyway
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import javax.sql.DataSource

@Service
@Transactional
class MigrationService(
//    private val shardDatabaseProperties: ShardDatabaseProperties,
    private val shardDataSources: Map<String, DataSource>,
) {

    companion object {
        const val MIGRATION_LOCATION = "classpath:/db/migration"
    }

    fun migrateAllShards(): Map<String, String> {
        val results = mutableMapOf<String, String>()
        shardDataSources.forEach { (shardName, datasource) ->
            try {
                val flyway = Flyway.configure()
                    .dataSource(datasource)
                    .locations(MIGRATION_LOCATION)
                    .baselineOnMigrate(true)
                    .load()

                flyway.migrate()
                results[shardName] = "SUCCESS"
            } catch (e: Exception) {
                results[shardName] = "FAILED: ${e.message}"
            }
        }
        return results
    }

    fun migrateSingleShard(shardName: String): String {
        val dataSourceProperties = shardDataSources[shardName]
            ?: throw IllegalArgumentException("Shard '$shardName' not found")

        try {
            val flyway = Flyway.configure()
                .dataSource(dataSourceProperties)
                .locations(MIGRATION_LOCATION)
                .baselineOnMigrate(true)
                .load()

            flyway.migrate()
            return "SUCCESS"
        } catch (e: Exception) {
            return "FAILED: ${e.message}"
        }
    }

    fun getMigrationInfoForAll(): List<Map<String, Any?>> {
        return shardDataSources.map { (_, dataSourceProperties) ->
            val flyway = Flyway.configure()
                .dataSource(dataSourceProperties)
                .locations(MIGRATION_LOCATION)
                .load()

            mapOf(
                "currentVersion" to flyway.info()?.current()?.version,
                "pending" to flyway.info()?.pending()?.map { it.version },
                "applied" to flyway.info()?.applied()?.map { it.version }
            )
        }
    }

    fun getMigrationInfo(shardName: String): Map<String, Any?> {
        val dataSourceProperties = shardDataSources[shardName]
            ?: throw IllegalArgumentException("Shard '$shardName' not found")

        val flyway = Flyway.configure()
            .dataSource(dataSourceProperties)
            .locations(MIGRATION_LOCATION)
            .load()

        return mapOf(
            "currentVersion" to flyway.info().current().version,
            "pending" to flyway.info().pending().map { it.version },
            "applied" to flyway.info().applied().map { it.version }
        )
    }
}
