package com.example.uuidgenerator.resource;

import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.Matchers.matchesPattern;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

@QuarkusTest
public class UuidResourceTest {

    @Test
    public void testGenerateUuidEndpoint() {
        given()
            .when().get("/api/uuids/generate")
            .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("uuid", notNullValue())
                .body("uuid", matchesPattern("^[0-9a-f]{8}-[0-9a-f]{4}-1[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"));
    }

    @Test
    public void testGeneratedUuidIsVersion1() {
        given()
            .when().get("/api/uuids/generate")
            .then()
                .statusCode(200)
                .body("uuid", matchesPattern("^[0-9a-f]{8}-[0-9a-f]{4}-1[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"));
    }

    @Test
    public void testMultipleCallsGenerateDifferentUuids() {
        String uuid1 = given()
            .when().get("/api/uuids/generate")
            .then()
                .statusCode(200)
                .extract().path("uuid");
        
        String uuid2 = given()
            .when().get("/api/uuids/generate")
            .then()
                .statusCode(200)
                .extract().path("uuid");
        
        assertNotEquals(uuid1, uuid2, "Subsequent calls should generate different UUIDs");
    }
}
