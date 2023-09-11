FROM eclipse-temurin:17.0.8.1_1-jre-jammy

WORKDIR /app

COPY build/libs/k8s-test.jar ./app.jar

CMD [ "java", "-jar", "app.jar" ]
