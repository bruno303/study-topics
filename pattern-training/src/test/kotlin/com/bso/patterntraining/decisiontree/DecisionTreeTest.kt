package com.bso.patterntraining.decisiontree

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class DecisionTreeTest {
    @Test
    fun `should evaluate correctly when have only one condition`() {
        val subject = decisionTree<FakeInput, Int> {
            node {
                condition = { it.number == 8 }
                onApproved = { it.number * 2 }
            }
        }
        val result: Int = subject.evaluate(FakeInput(8))
        assertEquals(16, result)
    }

    @Test
    fun `should throw when no condition was configured`() {
        val subject = decisionTree<FakeInput, Int> {}
        val ex: IllegalStateException = assertThrows { subject.evaluate(FakeInput(8)) }
        assertEquals("No condition was configured", ex.message)
    }

    @Test
    fun `should evaluate correctly when have multiple conditions`() {
        val subject = decisionTree<FakeInput, Int> {
            (1..10).forEach { index ->
                node {
                    condition = { it.number == index }
                    onApproved = { it.number * index }
                }
            }
        }
        (10 downTo 1).forEach {
            assertEquals(it * it, subject.evaluate(FakeInput(it)))
        }
    }

    @Test
    fun `should call default action when no condition was accepted`() {
        val subject = decisionTree<FakeInput, Int> {
            (1..10).forEach { index ->
                node {
                    condition = { it.number == index }
                    onApproved = { it.number * index }
                }
            }
            default { it.number * -1 }
        }
        val result = subject.evaluate(FakeInput(15))
        assertEquals(-15, result)
    }

    @Test
    fun `should throw when no condition was accepted and no default action configured`() {
        val subject = decisionTree<FakeInput, Int> {
            (1..10).forEach { index ->
                node {
                    condition = { it.number == index }
                    onApproved = { it.number * index }
                }
            }
        }
        val ex: IllegalStateException = assertThrows { subject.evaluate(FakeInput(15)) }
        assertEquals("None of the conditions was true", ex.message)
    }
}

data class FakeInput(
    val number: Int
)