package com.bso.postgressharding.web.controllers

import com.bso.postgressharding.application.migration.MigrationService
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*

@RestController
@RequestMapping("/api/migrations")
class MigrationController(
    private val migrationService: MigrationService
) {

    @PostMapping("/migrate/all")
    fun migrateAllShards(): ResponseEntity<Map<String, String>> {
        val results = migrationService.migrateAllShards()
        return ResponseEntity.ok(results)
    }

    @PostMapping("/migrate/{shardName}")
    fun migrateSingleShard(@PathVariable shardName: String): ResponseEntity<String> {
        val result = migrationService.migrateSingleShard(shardName)
        return ResponseEntity.ok(result)
    }

    @GetMapping("/info/{shardName}")
    fun getMigrationInfo(@PathVariable shardName: String): ResponseEntity<Map<String, Any?>> {
        return ResponseEntity.ok(migrationService.getMigrationInfo(shardName))
    }

    @GetMapping("/info/all")
    fun getAllShardsInfo(): ResponseEntity<List<Map<String, Any?>>> {
        val results = migrationService.getMigrationInfoForAll()
        return ResponseEntity.ok(results)
    }
}
