package com.bso.service1.config

import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.http.HttpMethod
import org.springframework.security.authorization.AuthorityAuthorizationManager
import org.springframework.security.config.annotation.web.builders.HttpSecurity
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity
import org.springframework.security.config.http.SessionCreationPolicy
import org.springframework.security.config.web.servlet.invoke
import org.springframework.security.core.GrantedAuthority
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.oauth2.jwt.Jwt
import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationConverter
import org.springframework.security.web.SecurityFilterChain

/**
 * Some help to understand keycloak and spring:
 * https://stackoverflow.com/questions/73008055/why-do-we-need-keycloak-permissions-policies-scopes-if-we-want-to-control-access
 */
@Configuration
@EnableWebSecurity
class EndpointSecurityConfig {
    private val logger: Logger by lazy { LoggerFactory.getLogger(this::class.java) }

    /**
     * hasAuthority -> scope defined in keycloak
     */
    @Bean
    fun filterChain2KotlinLike(http: HttpSecurity): SecurityFilterChain {
        http {
            cors { disable() }
            csrf { disable() }
            authorizeRequests {
                authorize(
                    HttpMethod.GET,
                    "/hello/protected",
                    hasAnyAuthority(internal, humanResources, qa)
                )
                authorize(HttpMethod.GET, "/hello/denied", denyAll)
                authorize(HttpMethod.GET, "/hello/unprotected", permitAll)
            }
            oauth2ResourceServer {
                jwt {
                    jwtAuthenticationConverter = jwtAuthenticationConverter()
                }
            }
            sessionManagement {
                sessionCreationPolicy = SessionCreationPolicy.STATELESS
            }
        }
        return http.build()
    }

    private fun jwtAuthenticationConverter(): JwtAuthenticationConverter {
        return JwtAuthenticationConverter().apply {
            setJwtGrantedAuthoritiesConverter { jwt ->
                logger.info("Roles: {}", jwt.getRoles())
                logger.info("Scopes: {}", jwt.getAuthorities())
                mutableListOf<GrantedAuthority>().apply {
                    addAll(jwt.getRoles())
                    addAll(jwt.getAuthorities())
                }
            }
        }
    }
}

private fun Jwt.getAuthorities(): List<SimpleGrantedAuthority> {
    val scopes = (claims["scope"] ?: return listOf()) as String
    if (scopes == "") return listOf()
    return scopes
        .split(" ")
        .map { authority -> SimpleGrantedAuthority(authority) }
}

@Suppress("UNCHECKED_CAST")
private fun Jwt.getRoles(): List<SimpleGrantedAuthority> {
    return ((claims["roles"] ?: return listOf()) as List<String>)
        .map { role -> SimpleGrantedAuthority("ROLE_$role") }
}