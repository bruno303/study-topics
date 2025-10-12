package com.example.uuidgenerator.service;

import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.Test;

import jakarta.inject.Inject;
import java.util.UUID;
import java.util.regex.Pattern;

import static org.junit.jupiter.api.Assertions.*;

@QuarkusTest
public class UuidServiceTest {

    private static final Pattern UUID_PATTERN = 
        Pattern.compile("^[0-9a-f]{8}-[0-9a-f]{4}-1[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$");

    @Inject
    UuidService uuidService;

    @Test
    public void testGenerateTimeOrderedUuid() {
        // When
        String uuid = uuidService.generateTimeOrderedUuid();
        
        // Then
        assertNotNull(uuid, "Generated UUID should not be null");
        assertTrue(UUID_PATTERN.matcher(uuid).matches(), "Generated UUID should match the expected pattern");
        
        // Verify it's a version 1 (time-based) UUID
        UUID uuidObj = UUID.fromString(uuid);
        assertEquals(1, uuidObj.version(), "Generated UUID should be version 1 (time-based)");
    }

    @Test
    public void testGenerateTimeOrderedUuidWithCustomNode() {
        // Given
        byte[] nodeId = new byte[]{0x01, 0x23, 0x45, 0x67, (byte) 0x89, (byte) 0xAB};
        
        // When
        String uuid = uuidService.generateTimeOrderedUuid(nodeId);
        
        // Then
        assertNotNull(uuid, "Generated UUID with custom node should not be null");
        assertTrue(UUID_PATTERN.matcher(uuid).matches(), "Generated UUID should match the expected pattern");
        
        // Verify it's a version 1 (time-based) UUID
        UUID uuidObj = UUID.fromString(uuid);
        assertEquals(1, uuidObj.version(), "Generated UUID should be version 1 (time-based)");
    }

    @Test
    public void testGenerateMultipleUuidsAreUnique() {
        // When
        String uuid1 = uuidService.generateTimeOrderedUuid();
        String uuid2 = uuidService.generateTimeOrderedUuid();
        
        // Then
        assertNotEquals(uuid1, uuid2, "Subsequent calls should generate different UUIDs");

        // Verify uuid1 is lower or comes before uuid2
        int compareResult = uuid1.compareTo(uuid2);
        assertTrue(compareResult <= 0, "uuid1 should be lower or come before uuid2");
    }
}
