package com.bso.postgressharding.infra.repository.impl

import com.bso.postgressharding.application.user.UserRepository
import com.bso.postgressharding.domain.entities.User
import com.bso.postgressharding.infra.repository.ConsistentHashRouter
import jakarta.persistence.EntityManagerFactory
import org.springframework.data.jpa.repository.support.JpaRepositoryFactory
import org.springframework.orm.jpa.JpaTransactionManager
import org.springframework.orm.jpa.LocalContainerEntityManagerFactoryBean
import org.springframework.orm.jpa.vendor.HibernateJpaVendorAdapter
import org.springframework.stereotype.Component
import org.springframework.transaction.support.TransactionTemplate
import java.util.UUID
import javax.sql.DataSource
import kotlin.jvm.optionals.getOrNull

typealias PersistenceUser = com.bso.postgressharding.infra.entities.User

@Component
class UserRepositoryJpaImpl(
    shardDataSources: Map<String, DataSource>
) : UserRepository {

    private val router = ConsistentHashRouter(shardDataSources.keys.toList())
    private val userRepositories: Map<String, UserRepositoryJpa> = shardDataSources.mapValues { (_, ds) ->
        val emFactory: EntityManagerFactory = LocalContainerEntityManagerFactoryBean().apply {
            val vendorAdapter = HibernateJpaVendorAdapter().apply {
                setGenerateDdl(true)
                setShowSql(true)
                setDatabasePlatform("org.hibernate.dialect.PostgreSQLDialect")
            }

            val jpaProperties = mapOf(
                "hibernate.hbm2ddl.auto" to "none",
                "hibernate.dialect" to "org.hibernate.dialect.PostgreSQLDialect",
                "jakarta.persistence.jdbc.url" to (dataSource?.connection?.metaData?.url)
            )

            dataSource = ds
            setPackagesToScan("com.bso.postgressharding.infra")
            jpaVendorAdapter = vendorAdapter
            setJpaPropertyMap(jpaProperties)
            afterPropertiesSet()
        }.nativeEntityManagerFactory

        JpaTransactionManager(emFactory).also { tm ->
            TransactionTemplate(tm).execute { }
        }

        JpaRepositoryFactory(emFactory.createEntityManager()).getRepository(UserRepositoryJpa::class.java)
    }

    private fun getShard(userId: String): UserRepositoryJpa {
        val shardName = router.getNode(userId)
        return userRepositories[shardName]!!
    }

    override fun save(user: User): User = getShard(user.id.toString())
        .saveAndFlush(user.toPersistence())
        .toDomain()

    override fun findById(id: UUID): User? = getShard(id.toString())
        .findById(id.toString())
        .getOrNull()?.toDomain()
}

private fun User.toPersistence(): PersistenceUser = PersistenceUser(
    this.id.toString(),
    this.name,
    this.email
)

private fun PersistenceUser.toDomain(): User = User(
    UUID.fromString(this.id),
    this.name,
    this.email
)