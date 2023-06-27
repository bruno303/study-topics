package com.bso.patterntraining.decisiontree

import kotlin.properties.Delegates

fun <T, R> decisionTree(
    configuration: DecisionTreeConfiguration<T, R>.() -> Unit
): DecisionTree<T, R> {
    val treeConfig = DecisionTreeConfiguration<T, R>()
        .apply(configuration)

    return treeConfig.nodeConfigs.map {
        DecisionTreeNode(
            condition = it.condition,
            onApproved = it.onApproved
        )
    }.let {
        DecisionTree(nodes = it, defaultAction = treeConfig.defaultAction)
    }
}

class DecisionTreeConfiguration<T, R> {
    private val configs: MutableList<DecisionTreeConditionConfiguration<T, R>> = mutableListOf()
    var defaultAction: ((T) -> R)? = null
        private set

    fun node(configuration: DecisionTreeConditionConfiguration<T, R>.() -> Unit) {
        val condition = DecisionTreeConditionConfiguration<T, R>().apply(configuration)
        configs.add(condition)
    }

    fun default(action: (T) -> R) {
        defaultAction = action
    }

    val nodeConfigs: List<DecisionTreeConditionConfiguration<T, R>>
        get() = configs.toList()
}

class DecisionTreeConditionConfiguration<T, R> {
    var condition: (T) -> Boolean by Delegates.notNull()
    var onApproved: (T) -> R by Delegates.notNull()
}