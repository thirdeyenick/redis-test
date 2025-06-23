package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func main() {
	redisServer := getEnv("REDIS_SERVER", "localhost:6379")
	redisUsername := getEnv("REDIS_USERNAME", "default")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	webserverPort := getEnv("PORT", "8080")
	useInsecureTLS := getEnv("REDIS_INSECURE_TLS", "false")
	useTLS, err := strconv.ParseBool(useInsecureTLS)
	if err != nil {
		log.Fatalf("can not parse env variable %q: %v", "REDIS_INSECURE_TLS", err)
	}

	var tlsConfig *tls.Config
	if useTLS {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:      redisServer,
		Username:  redisUsername,
		Password:  redisPassword, // no password by default
		DB:        0,             // use default DB
		TLSConfig: tlsConfig,
	})

	// Test Redis connection
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}
	log.Println("connected to Redis")

	http.HandleFunc("/", handleIndex)

	log.Printf("serving on http://localhost:%s ...", webserverPort)
	log.Fatal(http.ListenAndServe(":"+webserverPort, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Increment counter in Redis
	count, err := rdb.Incr(ctx, "visit_count").Result()
	if err != nil {
		http.Error(w, "Error accessing Redis", http.StatusInternalServerError)
		log.Printf("got error when increasing counter: %v", err)
		return
	}
	log.Printf("page visits: %d", count)

	// Simple HTML page
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head><title>Test</title></head>
		<body>
			<h1>all is working</h1>
			<p>nothing to see here</p>
		</body>
		</html>
	`)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
