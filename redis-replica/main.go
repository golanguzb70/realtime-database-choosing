package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdbMaster        *redis.Client
	replicas         *CustomLoadBalancer
	masterAddr             = "localhost:6379"
	replicaAddrs           = []string{"localhost:6380", "localhost:6381", "localhost:6382"}
	ctx                    = context.Background()
	lastDriverId     int64 = 0
	lastReadDriverId int64 = 0
	mu                     = &sync.Mutex{}
)

const (
	testCycleCount              = 1
	writeGoroutinesCount        = 1
	readGoroutinesCount         = 35
	numDrivers                  = 1_000_000
	writeOpsPerMinute           = 1000_000
	singleGetOpsPerMinute       = 1_000_000
	multiGetRadOpsPerMinute     = 1_500_000
	multiGetGeoHashOpsPerMinute = 500_000
)

func main() {
	// Connect to Redis
	rdbMaster = redis.NewClient(&redis.Options{
		Addr: masterAddr,
	})
	_, err := rdbMaster.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis master connection error:", err)
	}
	replicas, err = NewLoadBalancer(replicaAddrs)
	if err != nil {
		log.Fatal("Redis replicas connection error:", err)
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
	time.Sleep(time.Second * 20)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentSingleGets()
		fmt.Println("ConcurrentSingleGets Operation count - ", ops)
		mt.Lock()
		defer mt.Unlock()
		totalReadOperations += ops
		totalReadErrors += errCount
	}(wg)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentListGetInRaius()
		fmt.Println("ConcurrentListGetInRaius Operation count - ", ops)
		mu.Lock()
		defer mu.Unlock()
		totalReadOperations += ops
		totalReadErrors += errCount
	}(wg)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		ops, errCount := ConcurrentListInGeoHash()
		fmt.Println("ConcurrentListInGeoHash Operation count - ", ops)
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

type CustomLoadBalancer struct {
	clients []*redis.Client
	offset  int
}

func NewLoadBalancer(addrs []string) (*CustomLoadBalancer, error) {
	clients := []*redis.Client{}
	for _, addr := range addrs {
		client := redis.NewClient(&redis.Options{
			Addr: addr,
		})

		if _, err := client.Ping(ctx).Result(); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return &CustomLoadBalancer{
		clients: clients,
		offset:  0,
	}, nil
}

func (cl *CustomLoadBalancer) Get() *redis.Client {
	mu.Lock()
	defer mu.Unlock()

	if cl.offset >= len(cl.clients) {
		cl.offset = 0
	}

	client := cl.clients[cl.offset]
	cl.offset++
	return client
}
