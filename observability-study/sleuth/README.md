## Sleuth

This project is a test for distributed tracing with Sleuth.


## Flow diagram
```mermaid
graph LR
    Service1 -->|HTTP Request| Service2
    Service2 -->|SQS Queue| Service3
  
```

Communication between Service1 and Service2 is sync. This means that just use `spring-cloud-starter-sleuth` is enough to have a functional distributed tracing.

Communication between Service1 and Service2 is async (using SQS). In this scenario, in order to distributed tracing work, some code is necessary. This is done using a injector/extractor that reads/writes in SQS headers (map<string, string>). For Service3 use the received tracing, this is needed:
```java
try (var ws = tracer.withSpanInScope(span)) {
    // work here
}
```
Inside the `try-with-resources`, logs will have traceId and spanId correctly configured.


## References
* [json-logging](https://dev.to/anandsunderraman/json-logging-in-spring-boot-applications-2j33)
* [spring-boot logging](https://docs.spring.io/spring-boot/docs/current/reference/html/features.html#features.logging.custom-log-configuration)
* [logback layouts](https://logback.qos.ch/manual/layouts.html)