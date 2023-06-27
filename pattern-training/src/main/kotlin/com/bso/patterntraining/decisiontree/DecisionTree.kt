package com.bso.patterntraining.decisiontree

class DecisionTree<T, R>(
    private val nodes: List<DecisionTreeNode<T, R>>,
    private val defaultAction: ((T) -> R)? = null
) {
    fun evaluate(input: T): R {
        if (nodes.isEmpty()) {
            throw IllegalStateException("No condition was configured")
        }
        for (node in nodes) {
            if (node.condition(input)) {
                return node.onApproved(input)
            }
        }
        if (defaultAction == null) {
            throw IllegalStateException("None of the conditions was true")
        }
        return defaultAction.invoke(input)
    }
}

data class DecisionTreeNode<T, R>(
    val condition: (T) -> Boolean,
    val onApproved: (T) -> R
)