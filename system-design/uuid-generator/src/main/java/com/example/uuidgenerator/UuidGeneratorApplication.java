package com.example.uuidgenerator;

import io.quarkus.runtime.Quarkus;
import io.quarkus.runtime.annotations.QuarkusMain;

@QuarkusMain
public class UuidGeneratorApplication {
    public static void main(String[] args) {
        Quarkus.run(args);
    }
}
