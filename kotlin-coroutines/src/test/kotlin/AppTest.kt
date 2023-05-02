import kotlinx.coroutines.*
import org.junit.jupiter.api.Test
import java.time.Instant
import java.util.concurrent.Executors
import kotlin.test.assertEquals

class AppTest {

    @Test
    fun `test with runBlocking`() {
        val start = Instant.now()

        runBlocking {
            val query2 = async { slowQuery2() }
            val query1 = async { slowQuery() }

            val result = listOf(query1, query2)
                .awaitAll()
                .flatten()

            val end = Instant.now()

            assertEquals(10, result.size)
            println("RunBlocking executed in ${end.minusMillis(start.toEpochMilli()).toEpochMilli()} millis")
        }
    }

    @Test
    fun `test with withContext`() = runBlocking {
        val start = Instant.now()

        // withContext is just like call async and await in sequence.
        // can be replaced by           async { slowQuery2() }.await()
        val result2 = withContext(Dispatchers.Default) { slowQuery2() }
        val result1 = withContext(Dispatchers.Default) { slowQuery() }

        val result = mutableListOf<Int>().apply {
            addAll(result2)
            addAll(result1)
        }

        val end = Instant.now()

        assertEquals(10, result.size)
        println("WithContext executed in ${end.minusMillis(start.toEpochMilli()).toEpochMilli()} millis")
    }

    @Test
    fun `test with coroutineScope`() = runBlocking {
        val start = Instant.now()

        // coroutineScope is another builder function to create an new scope to run coroutines.
        // can be only used in suspend methods or inside another CoroutineScope
        // everything inside a coroutineScope will be managed by this scope.
        //
        // If some task fail, other tasks in the same scope will be canceled.
        // This provides a good default control over tasks running async
        coroutineScope {
            val query2 = async { slowQuery2() }
            val query1 = async { slowQuery() }

            val result = listOf(query1, query2)
                .awaitAll()
                .flatten()

            val end = Instant.now()

            assertEquals(10, result.size)
            println("WithContext executed in ${end.minusMillis(start.toEpochMilli()).toEpochMilli()} millis")
        }
    }

    @Test
    fun `test with custom Dispatcher (threadpool)`() = runBlocking {
        val dispatcher = Executors.newFixedThreadPool(4).asCoroutineDispatcher()

        val start = Instant.now()

        // a dispatcher is the threadpool where the task will be executed. We have a few default in Dispatchers class
        // but we can create our own using Executors class from java
        val query2 = async(dispatcher) { slowQuery2() }
        val query1 = async(dispatcher) { slowQuery() }

        val result = listOf(query1, query2)
            .awaitAll()
            .flatten()

        val end = Instant.now()
        dispatcher.close()

        assertEquals(10, result.size)
        println("WithContext executed in ${end.minusMillis(start.toEpochMilli()).toEpochMilli()} millis")
    }

    // this function is suspend because of delay function. Delay() wait for the millis passed,
    // but don't block the thread.
    private suspend fun slowQuery(): List<Int> {
        println("Starting query 1")
        delay(1000)
        println("Returning query 1")
        return listOf(1, 2, 3, 4, 5)
    }

    // this function is suspend because of delay function. Delay() wait for the millis passed,
    // but don't block the thread.
    private suspend fun slowQuery2(): List<Int> {
        println("Starting query 2")
        delay(2000)
        println("Returning query 2")
        return listOf(6, 7, 8, 9, 10)
    }
}