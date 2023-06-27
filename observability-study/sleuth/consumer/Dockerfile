FROM eclipse-temurin:17.0.4.1_1-jdk AS builder

WORKDIR /app-build

# gradle
COPY gradlew gradlew
COPY gradle gradle

# project files
COPY settings.gradle settings.gradle
COPY build.gradle build.gradle
COPY src src

RUN ./gradlew build -x check -x test --no-daemon


FROM eclipse-temurin:17.0.4.1_1-jre AS runner

EXPOSE 8080
WORKDIR /app

COPY --from=builder /app-build/build/libs/sleuth-service2.jar app.jar
ENTRYPOINT [ "java", "-jar", "app.jar" ]
