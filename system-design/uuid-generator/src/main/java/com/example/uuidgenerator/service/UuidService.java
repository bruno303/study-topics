package com.example.uuidgenerator.service;

import jakarta.enterprise.context.ApplicationScoped;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.security.SecureRandom;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicLong;

/**
 * Service for generating time-ordered UUIDs (version 1).
 * This implementation generates time-based UUIDs without external dependencies.
 */
@ApplicationScoped
public class UuidService {
    private static final Logger LOG = LoggerFactory.getLogger(UuidService.class);
    private static final String LOG_PREFIX = "[UUID-Service]";
    
    private static final long NUM_100NS_INTERVALS_SINCE_UUID_EPOCH = 0x01b21dd213814000L;
    private final byte[] nodeId;
    private final AtomicLong lastTime = new AtomicLong(0);
    private final SecureRandom secureRandom = new SecureRandom();
    private final AtomicLong clockSequence = new AtomicLong(secureRandom.nextInt() & 0x3FFF);
    
    public UuidService() {
        // Generate a random node ID if not specified
        this.nodeId = new byte[6];
        secureRandom.nextBytes(nodeId);
        
        // Set multicast bit to indicate this is a locally administered address
        this.nodeId[0] |= 0x01;
    }

    /**
     * Generates a new time-ordered UUID (version 1).
     * The generated UUIDs are time-ordered and include a node identifier.
     * 
     * @return a time-ordered UUID string
     */
    public String generateTimeOrderedUuid() {
        LOG.debug("{} Generating time-ordered UUID with default node ID", LOG_PREFIX);
        long startTime = System.nanoTime();
        try {
            String uuid = generateUuidV1(nodeId);
            LOG.debug("{} Successfully generated UUID in {} ns", LOG_PREFIX, (System.nanoTime() - startTime));
            return uuid;
        } catch (Exception e) {
            LOG.error("{} Error generating UUID: {}", LOG_PREFIX, e.getMessage(), e);
            throw e;
        }
    }

    /**
     * Generates a new time-ordered UUID with a custom node ID.
     * 
     * @param customNodeId the node ID to use (6 bytes)
     * @return a time-ordered UUID string
     * @throws IllegalArgumentException if the node ID is invalid
     */
    public String generateTimeOrderedUuid(byte[] customNodeId) {
        LOG.debug("{} Generating time-ordered UUID with custom node ID", LOG_PREFIX);
        long startTime = System.nanoTime();
        
        if (customNodeId == null || customNodeId.length != 6) {
            String errorMsg = "Invalid node ID. Expected 6 bytes, got " + 
                           (customNodeId == null ? "null" : customNodeId.length + " bytes");
            LOG.error("{} {}", LOG_PREFIX, errorMsg);
            throw new IllegalArgumentException(errorMsg);
        }
        
        try {
            String uuid = generateUuidV1(customNodeId);
            LOG.debug("{} Successfully generated UUID with custom node in {} ns", LOG_PREFIX, 
                    (System.nanoTime() - startTime));
            return uuid;
        } catch (Exception e) {
            LOG.error("{} Error generating UUID with custom node: {}", LOG_PREFIX, e.getMessage(), e);
            throw e;
        }
    }
    
 private String generateUuidV1(byte[] node) {
        LOG.trace("{} Starting UUID v1 generation with node: {}", LOG_PREFIX, bytesToHex(node));
        long time = getCurrentTime();
        long clockSeq = clockSequence.get();
        LOG.trace("{} Using timestamp: {}, clock sequence: {}", LOG_PREFIX, time, clockSeq);
        
        // Convert to UUID format
        long msb = ((time & 0x0fff_0000_0000_0000L) >>> 48)
                 | ((time & 0x0000_ffff_0000_0000L) << 16)
                 | ((time & 0x0000_0000_ffff_ffffL) << 32)
                 | (1L << 12); // version 1
        
        long lsb = ((long) (clockSeq & 0x3FFF) << 48) // clock sequence
                 | 0x8000000000000000L // variant 2
                 | ((long) (node[0] & 0xFF) << 40)
                 | ((long) (node[1] & 0xFF) << 32)
                 | ((long) (node[2] & 0xFF) << 24)
                 | ((long) (node[3] & 0xFF) << 16)
                 | ((long) (node[4] & 0xFF) << 8)
                 | (node[5] & 0xFF);
        
        UUID uuid = new UUID(msb, lsb);
        LOG.trace("{} Generated UUID: {}", LOG_PREFIX, uuid);
        return uuid.toString();
    }
    
    private long getCurrentTime() {
        long time = (System.currentTimeMillis() * 10000) + NUM_100NS_INTERVALS_SINCE_UUID_EPOCH;
        LOG.trace("{} Generated timestamp: {}", LOG_PREFIX, time);
        
        // Handle clock sequence if we have a time collision
        while (true) {
            long last = lastTime.get();
            if (time > last) {
                if (lastTime.compareAndSet(last, time)) {
                    break;
                }
            } else {
                // Clock went backwards or same timestamp, increment clock sequence
                time = last + 1;
                long newClockSeq = clockSequence.incrementAndGet();
                LOG.warn("{} Clock sequence incremented due to timestamp collision. New sequence: {}", 
                        LOG_PREFIX, newClockSeq);
            }
        }
        
        return time;
    }
    
    /**
     * Converts a byte array to a hexadecimal string for logging purposes.
     * 
     * @param bytes the byte array to convert
     * @return a hexadecimal string representation of the byte array
     */
    private String bytesToHex(byte[] bytes) {
        if (bytes == null) return "null";
        StringBuilder sb = new StringBuilder(bytes.length * 2);
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}
