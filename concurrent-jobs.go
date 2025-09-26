package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func ConcurrentUpdates() (int, int) {
	// fmt.Println("Starting concurrent updates test...")

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Channel to collect statistics
	statsChan := make(chan Stats, writeGoroutinesCount*testCycleCount)

	// Start the test cycles
	for cycle := range testCycleCount {
		fmt.Printf("Starting Create/Update test cycle %d/%d\n", cycle+1, testCycleCount)

		// Launch concurrent goroutines for this cycle
		for i := range writeGoroutinesCount {
			wg.Add(1)
			go func(workerID, cycleID int) {
				defer wg.Done()

				// Run operations for 1 minute
				startTime := time.Now()
				opsPerWorker := (writeOpsPerMinute / writeGoroutinesCount) + writeOpsPerMinute%writeGoroutinesCount
				operationCount := 0
				errorCount := 0

				for time.Since(startTime) < time.Minute {
					// Generate a random driver ID
					driverID := getNextDriverId()

					// Generate fake driver data
					driver := GenerateFakeDriver(driverID)

					// Update the driver in Redis
					err := UpsertDriver(driver)
					if err != nil {
						errorCount++
						log.Printf("Worker %d: Error updating driver %d: %v", workerID, driverID, err)
					} else {
						operationCount++
					}
					if operationCount >= opsPerWorker {
						break
					}
				}

				// Send statistics
				statsChan <- Stats{
					WorkerID:   workerID,
					CycleID:    cycleID,
					Operations: operationCount,
					Errors:     errorCount,
					Duration:   time.Since(startTime),
				}
			}(i, cycle)
		}

		// Wait for this cycle to complete
		wg.Wait()

		// Small delay between cycles
		time.Sleep(time.Second * 2)
	}

	// Close the stats channel
	close(statsChan)

	// Collect and analyze statistics
	return analyzeUpdateStats(statsChan)
}

func ConcurrentSingleGets() (int, int) {
	// fmt.Println("Starting concurrent single GETs test...")

	var wg sync.WaitGroup
	statsChan := make(chan Stats, readGoroutinesCount*testCycleCount)

	for cycle := 0; cycle < testCycleCount; cycle++ {
		fmt.Printf("Starting Single GET test cycle %d/%d\n", cycle+1, testCycleCount)

		for i := 0; i < readGoroutinesCount; i++ {
			wg.Add(1)
			go func(workerID, cycleID int) {
				defer wg.Done()

				startTime := time.Now()
				opsPerWorker := (singleGetOpsPerMinute / readGoroutinesCount) + singleGetOpsPerMinute%readGoroutinesCount
				operationCount := 0
				errorCount := 0

				for time.Since(startTime) < time.Minute {
					driverID := getNextDriverIdRead()
					_, err := GetDriver(driverID)
					if err != nil {
						errorCount++
						log.Printf("Worker %d: Error getting driver %d: %v", workerID, driverID, err)
					} else {
						operationCount++
					}
					if operationCount >= opsPerWorker {
						time.Sleep(time.Until(startTime.Add(time.Minute)))
					}
				}

				statsChan <- Stats{
					WorkerID:   workerID,
					CycleID:    cycleID,
					Operations: operationCount,
					Errors:     errorCount,
					Duration:   time.Since(startTime),
				}
			}(i, cycle)
		}

		wg.Wait()
		time.Sleep(time.Second * 2)
	}

	close(statsChan)
	return analyzeUpdateStats(statsChan)
}

func ConcurrentListGetInRaius() (int, int) {
	// fmt.Println("Starting concurrent list GETs in radius test...")

	var wg sync.WaitGroup
	statsChan := make(chan Stats, readGoroutinesCount*testCycleCount)

	for cycle := 0; cycle < testCycleCount; cycle++ {
		fmt.Printf("Starting List GET in Radius test cycle %d/%d\n", cycle+1, testCycleCount)

		for i := 0; i < readGoroutinesCount; i++ {
			wg.Add(1)
			go func(workerID, cycleID int) {
				defer wg.Done()

				startTime := time.Now()
				opsPerWorker := (multiGetRadOpsPerMinute / readGoroutinesCount) + multiGetRadOpsPerMinute%readGoroutinesCount

				operationCount := 0
				errorCount := 0

				for time.Since(startTime) < time.Minute {
					lat, lng, _ := GetRandomLatLong()
					_, err := GetDriverInRadius(Location{Lat: lat, Long: lng}, 5, 30) // 5km radius
					if err != nil {
						errorCount++
						log.Printf("Worker %d: Error getting drivers in radius: %v", workerID, err)
					} else {
						operationCount++
					}
					if operationCount >= opsPerWorker {
						time.Sleep(time.Until(startTime.Add(time.Minute)))
					}
				}

				statsChan <- Stats{
					WorkerID:   workerID,
					CycleID:    cycleID,
					Operations: operationCount,
					Errors:     errorCount,
					Duration:   time.Since(startTime),
				}
			}(i, cycle)
		}

		wg.Wait()
		time.Sleep(time.Second * 2)
	}

	close(statsChan)
	return analyzeUpdateStats(statsChan)
}

func ConcurrentListInGeoHash() (int, int) {
	// fmt.Println("Starting concurrent list GETs in geohash test...")

	var wg sync.WaitGroup
	statsChan := make(chan Stats, readGoroutinesCount*testCycleCount)

	for cycle := 0; cycle < testCycleCount; cycle++ {
		fmt.Printf("Starting List GET in Geohash test cycle %d/%d\n", cycle+1, testCycleCount)

		for i := 0; i < readGoroutinesCount; i++ {
			wg.Add(1)
			go func(workerID, cycleID int) {
				defer wg.Done()

				startTime := time.Now()
				opsPerWorker := (multiGetGeoHashOpsPerMinute / readGoroutinesCount) + multiGetGeoHashOpsPerMinute%readGoroutinesCount

				operationCount := 0
				errorCount := 0

				for time.Since(startTime) < time.Minute {
					_, _, geohash := GetRandomLatLong()
					_, err := GetDriverForOrder(geohash, GetRandomTariffs(), 5)
					if err != nil {
						errorCount++
						log.Printf("Worker %d: Error getting drivers in geohash: %v", workerID, err)
					} else {
						operationCount++
					}
					if operationCount >= opsPerWorker {
						time.Sleep(time.Until(startTime.Add(time.Minute)))
					}
				}

				statsChan <- Stats{
					WorkerID:   workerID,
					CycleID:    cycleID,
					Operations: operationCount,
					Errors:     errorCount,
					Duration:   time.Since(startTime),
				}
			}(i, cycle)
		}

		wg.Wait()
		time.Sleep(time.Second * 2)
	}

	close(statsChan)
	return analyzeUpdateStats(statsChan)
}

// UpdateStats represents statistics for update operations
type Stats struct {
	WorkerID   int
	CycleID    int
	Operations int
	Errors     int
	Duration   time.Duration
}

func analyzeUpdateStats(statsChan <-chan Stats) (int, int) {
	var totalOps, totalErrors int

	for stats := range statsChan {
		totalOps += stats.Operations
		totalErrors += stats.Errors
	}

	return totalOps, totalErrors
}

func getNextDriverId() int64 {
	mu.Lock()
	defer mu.Unlock()
	if lastDriverId >= numDrivers {
		lastDriverId = 0
	}
	lastDriverId++
	return lastDriverId
}

func getNextDriverIdRead() int64 {
	mu.Lock()
	defer mu.Unlock()
	if lastReadDriverId >= numDrivers {
		lastReadDriverId = 0
	}
	lastReadDriverId++
	return lastReadDriverId
}
