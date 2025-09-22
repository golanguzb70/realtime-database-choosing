package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/pierrre/geohash"
)

func GenerateFakeDriver(id int64) Driver {
	// Seed the random number generator

	// Generate random location within a reasonable range (e.g., around a city center)
	// Using Tashkent, Uzbekistan as a reference point
	lat, lng, geoHash := GetRandomLatLong()
	location := Location{
		Lat:  lat,
		Long: lng,
	}

	// Generate random score (0-100)
	score := rand.Int63n(101)

	// Generate random phone charge percentage (0-100)
	charge := rand.Int63n(101)

	// Randomly set active status (80% chance of being active)
	active := rand.Float64() < 0.8

	// Generate last updated time (within last 24 hours)
	lastUpdated := time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour)
	lastUpdatedTime := strconv.FormatInt(lastUpdated.Unix(), 10)

	return Driver{
		Id:              id,
		GeoHash:         geoHash,
		Location:        location,
		ActiveTariffs:   GetRandomTariffs(),
		Score:           score,
		Charge:          charge,
		Active:          active,
		LastUpdatedTime: lastUpdatedTime,
	}
}

func GetRandomLatLong() (float64, float64, string) {
	baseLat := 41.2995
	baseLng := 69.2401

	// Add random offset within ~2000km radius
	latOffset := (rand.Float64() - 0.5) * 20.0 // ~1000km in each direction
	lngOffset := (rand.Float64() - 0.5) * 20.0
	lat := baseLat + latOffset
	lng := baseLng + lngOffset

	geoHash := geohash.Encode(lat, lng, 10)

	return lat, lng, geoHash
}

func GetRandomTariffs() []string {
	allTariffs := []string{"start", "comfort", "comfort+", "business", "premium"}
	numTariffs := rand.Intn(2) + 1 // 1 to 2 tariffs
	selectedTariffs := make([]string, numTariffs)

	// Shuffle and select tariffs
	shuffled := make([]string, len(allTariffs))
	copy(shuffled, allTariffs)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < numTariffs; i++ {
		selectedTariffs[i] = shuffled[i]
	}

	return selectedTariffs
}
