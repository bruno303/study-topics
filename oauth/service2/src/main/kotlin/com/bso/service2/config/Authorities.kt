package com.bso.service2.config

/**
 * Simple scope coming from jwt. Represents some individual permission like "internal" team members
 * or another "not user" permission
 */
typealias Authority = String

/**
 * Represent some professional level of a person. Or some users group, like "HR managers".
 */
typealias Role = String

// Roles
const val internal: Authority = "internal.all"

// Authorities
const val humanResources: Role = "HUMAN_RESOURCES"
const val qa: Role = "qa"