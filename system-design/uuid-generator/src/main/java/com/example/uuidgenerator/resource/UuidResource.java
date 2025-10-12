package com.example.uuidgenerator.resource;

import com.example.uuidgenerator.service.UuidService;
import jakarta.inject.Inject;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;

/**
 * REST endpoint for generating time-ordered UUIDs.
 */
@Path("/api/uuids")
public class UuidResource {

    @Inject
    UuidService uuidService;

    /**
     * Generates a new time-ordered UUID.
     * 
     * @return a JSON object containing the generated UUID
     */
    @GET
    @Path("/generate")
    @Produces(MediaType.APPLICATION_JSON)
    public UuidResponse generateUuid() {
        String uuid = uuidService.generateTimeOrderedUuid();
        return new UuidResponse(uuid);
    }

    /**
     * Simple DTO for the UUID response.
     */
    public record UuidResponse(String uuid) {
        public UuidResponse(String uuid) {
            this.uuid = uuid;
        }

        @Override
        public String uuid() {
            return uuid;
        }
    }
}
