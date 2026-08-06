package com.example

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController

@SpringBootApplication
class Application

@RestController
class HealthController {
    @GetMapping("/health")
    fun health() = "spring-kotlin-gradle"
}

fun main(args: Array<String>) {
    runApplication<Application>(*args)
}
