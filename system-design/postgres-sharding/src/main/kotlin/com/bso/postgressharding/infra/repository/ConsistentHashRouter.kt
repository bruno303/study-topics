package com.bso.postgressharding.infra.repository

class ConsistentHashRouter<T>(nodes: List<T>) {
    private val ring = sortedMapOf<Int, T>()

    init {
        nodes.forEach { node ->
            val hash = node.hashCode()
            ring[hash] = node
        }
    }

    fun getNode(key: String): T {
        val hash = key.hashCode()
        val tailMap = ring.tailMap(hash)
        return if (tailMap.isNotEmpty()) {
            tailMap.firstEntry().value
        } else {
            ring.firstEntry().value
        }
    }
}
