package com.bso.service2.config

import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.http.HttpMethod
import org.springframework.security.config.annotation.web.builders.HttpSecurity
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity
import org.springframework.security.config.http.SessionCreationPolicy
import org.springframework.security.config.web.servlet.invoke
import org.springframework.security.core.GrantedAuthority
import org.springframework.security.core.authority.SimpleGrantedAuthority
import org.springframework.security.oauth2.jwt.Jwt
import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationConverter
import org.springframework.security.web.SecurityFilterChain

@Configuration
@EnableWebSecurity
class EndpointSecurityConfig {
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
                    config(
                        withAuthorities(internal),
                        withRoles(humanResources, qa)
                    )
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
        return JwtAuthenticationConverter().also {
            it.setJwtGrantedAuthoritiesConverter { jwt ->
                val authorities: MutableCollection<GrantedAuthority> = mutableListOf()
                authorities.addAll(jwt.getRoles())
                authorities.addAll(jwt.getAuthorities())
                authorities
            }
        }
    }

    private fun withAuthorities(vararg authorities: Authority): String =
        "hasAnyAuthority('${authorities.joinToString("','")}')"

    private fun withRoles(vararg roles: Role): String =
        "hasAnyRole('${roles.joinToString("','")}')"

    private fun config(vararg permissions: String): String {
        return permissions.joinToString(" or ")
    }
}

@Suppress("UNCHECKED_CAST")
private fun Jwt.getAuthorities(): List<SimpleGrantedAuthority> {
    val scopes = (claims["scope"] ?: return listOf()) as String
    if (scopes == "") return listOf()
    return scopes
        .split(" ")
        .map { authority -> SimpleGrantedAuthority(authority) }
}

@Suppress("UNCHECKED_CAST")
private fun Jwt.getRoles(): List<SimpleGrantedAuthority> {
    val realmAccess = (claims["realm_access"] ?: return listOf()) as Map<*, *>
    return (realmAccess["roles"] as List<String>)
        .map { role -> SimpleGrantedAuthority("ROLE_$role") }
}