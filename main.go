package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client
var lastDriverId int64 = 0
var lastReadDriverId int64 = 0

var mu sync.Mutex

const (
	testCycleCount              = 1
	writeGoroutinesCount        = 20
	readGoroutinesCount         = 25
	numDrivers                  = 1_000_000
	writeOpsPerMinute           = 1_000_000
	singleGetOpsPerMinute       = 1_000_000
	multiGetRadOpsPerMinute     = 1_500_000
	multiGetGeoHashOpsPerMinute = 500_000
)

func main() {
	// Connect to Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal("Redis connection error:", err)
	}
	// Flush all existing data
	fmt.Println("Flushing all existing data...")
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		log.Fatalf("Failed to flush database: %v", err)
	}
	time.Sleep(time.Second * 2)
	fmt.Println("Database flushed successfully.")

	fmt.Println("Creating index...")
	_, err := rdb.Do(ctx, "FT.CREATE", "index", "ON", "HASH", "PREFIX", "1", "driver:", "SCHEMA",
		"driver_id", "NUMERIC", "SORTABLE",
		"location", "GEO",
		"geo_hash", "TEXT",
		"active_tariffs", "TAG", "SEPARATOR", "|",
		"score", "NUMERIC", "SORTABLE",
		"active", "TAG",
		"phone_charge_percent", "NUMERIC", "NOINDEX",
		"last_updated_time", "NUMERIC", "NOINDEX",
	).Result()
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	var (
		totalWriteOperations int
		totalWriteErrors     int
		totalReadOperations  int
		totalReadErrors      int
	)
	wg := &sync.WaitGroup{}
	mt := &sync.Mutex{}
	wg.Add(4)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		totalWriteOperations, totalWriteErrors = ConcurrentUpdates()
	}(wg)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentSingleGets()
		mt.Lock()
		defer mt.Unlock()
		totalReadOperations += ops
		totalReadErrors += errCount
	}(wg)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentListGetInRaius()
		mu.Lock()
		defer mu.Unlock()
		totalReadOperations += ops
		totalReadErrors += errCount
	}(wg)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentListInGeoHash()
		mu.Lock()
		defer mu.Unlock()
		totalReadOperations += ops
		totalReadErrors += errCount
	}(wg)

	wg.Wait()

	fmt.Println("\n|===== Summary of Write operations =====|")
	fmt.Printf("Total Write Operations: %d\n", totalWriteOperations)
	fmt.Printf("Total Write Operations per minute: %d\n", totalWriteOperations/testCycleCount)
	fmt.Printf("Total Write Errors: %d\n", totalWriteErrors)
	fmt.Println("\n|===== Summary of Read operations =====|")
	fmt.Printf("Total Read Operations: %d\n", totalReadOperations)
	fmt.Printf("Total Read Operations per minute: %d\n", totalReadOperations/testCycleCount)
	fmt.Printf("Total Read Errors: %d\n", totalReadErrors)
}
