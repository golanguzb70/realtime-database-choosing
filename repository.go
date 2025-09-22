package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Location struct {
	Lat  float64
	Long float64
}

var ActiveTariffs []string = []string{"start", "camfort|camfort+", "business"}

type Driver struct {
	Id              int64
	GeoHash         string
	Location        Location
	ActiveTariffs   []string
	Score           int64
	Charge          int64
	Active          bool
	LastUpdatedTime string
}

// Create driver if not exists if exists update it.
func UpsertDriver(in Driver) error {
	key := fmt.Sprintf("driver:%d", in.Id)

	// Prepare the hash fields
	fields := map[string]interface{}{
		"driver_id":            in.Id,
		"location":             fmt.Sprintf("%f,%f", in.Location.Lat, in.Location.Long),
		"geo_hash":             in.GeoHash,
		"active_tariffs":       strings.Join(in.ActiveTariffs, "|"),
		"score":                in.Score,
		"active":               fmt.Sprintf("%t", in.Active),
		"phone_charge_percent": in.Charge,
		"last_updated_time":    in.LastUpdatedTime,
	}

	// Use HSet to create or update the driver
	return rdb.HSet(ctx, key, fields).Err()
}

func GetDriver(id int64) (Driver, error) {
	key := fmt.Sprintf("driver:%d", id)

	// Get all fields from the hash
	result, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return Driver{}, err
	}

	if len(result) == 0 {
		return Driver{}, nil
	}

	// Parse the result into Driver struct
	driver := Driver{}

	if driverId, ok := result["driver_id"]; ok {
		if id, err := strconv.ParseInt(driverId, 10, 64); err == nil {
			driver.Id = id
		}
	}

	if location, ok := result["location"]; ok {
		parts := strings.Split(location, ",")
		if len(parts) == 2 {
			if lat, err := strconv.ParseFloat(parts[0], 64); err == nil {
				driver.Location.Lat = lat
			}
			if lng, err := strconv.ParseFloat(parts[1], 64); err == nil {
				driver.Location.Long = lng
			}
		}
	}

	if geoHash, ok := result["geo_hash"]; ok {
		driver.GeoHash = geoHash
	}

	if activeTariffs, ok := result["active_tariffs"]; ok {
		driver.ActiveTariffs = strings.Split(activeTariffs, "|")
	}

	if score, ok := result["score"]; ok {
		if s, err := strconv.ParseInt(score, 10, 64); err == nil {
			driver.Score = s
		}
	}

	if active, ok := result["active"]; ok {
		driver.Active = active == "true"
	}

	if charge, ok := result["phone_charge_percent"]; ok {
		if c, err := strconv.ParseInt(charge, 10, 64); err == nil {
			driver.Charge = c
		}
	}

	if lastUpdated, ok := result["last_updated_time"]; ok {
		driver.LastUpdatedTime = lastUpdated
	}

	return driver, nil
}

// In response sort by driver_id field
func GetDriverInRadius(location Location, radiusKm float64, limit int) ([]Driver, error) {
	// Build the search query for geospatial search
	query := fmt.Sprintf("@location:[%f %f %f km]", location.Long, location.Lat, radiusKm)

	// Execute the search with sorting by driver_id
	searchResult, err := rdb.Do(ctx, "FT.SEARCH", "index", query, "SORTBY", "driver_id", "ASC", "LIMIT", 0, limit).Result()
	if err != nil {
		return nil, err
	}

	// Parse the search result
	results, ok := searchResult.([]interface{})
	if !ok || len(results) < 2 {
		return []Driver{}, nil
	}

	// Skip the first element (total count) and process driver data
	driverData := results[1:]
	drivers := make([]Driver, 0, len(driverData))

	for i := 0; i < len(driverData); i += 2 {
		if i+1 >= len(driverData) {
			break
		}

		// Get the driver ID from the key
		key, ok := driverData[i].(string)
		if !ok {
			continue
		}

		// Extract driver ID from key (format: driver:123)
		keyParts := strings.Split(key, ":")
		if len(keyParts) != 2 {
			continue
		}

		driverId, err := strconv.ParseInt(keyParts[1], 10, 64)
		if err != nil {
			continue
		}

		// Get the driver details
		driver, err := GetDriver(driverId)
		if err != nil {
			continue
		}

		drivers = append(drivers, driver)
	}

	return drivers, nil
}

// In response sort by score field
func GetDriverForOrder(geoHash string, tariffs []string, limit int) ([]Driver, error) {
	// Build the search query
	var queryParts []string

	// Add geo_hash filter if provided
	if geoHash != "" {
		queryParts = append(queryParts, fmt.Sprintf("@geo_hash:%s*", geoHash))
	}

	// Add active filter
	queryParts = append(queryParts, "@active:{true}")
	queryParts = append(queryParts, fmt.Sprintf("@active_tariffs:{%s}", strings.Join(tariffs, "|")))

	query := strings.Join(queryParts, " ")

	// Execute the search with sorting by score (descending for best scores first)
	searchResult, err := rdb.Do(ctx, "FT.SEARCH", "index", query, "SORTBY", "score", "DESC", "LIMIT", 0, limit).Result()
	if err != nil {
		return nil, err
	}

	// Parse the search result
	results, ok := searchResult.([]interface{})
	if !ok || len(results) < 2 {
		return []Driver{}, nil
	}

	// Skip the first element (total count) and process driver data
	driverData := results[1:]
	drivers := make([]Driver, 0, len(driverData))

	for i := 0; i < len(driverData); i += 2 {
		if i+1 >= len(driverData) {
			break
		}

		// Get the driver ID from the key
		key, ok := driverData[i].(string)
		if !ok {
			continue
		}

		// Extract driver ID from key (format: driver:123)
		keyParts := strings.Split(key, ":")
		if len(keyParts) != 2 {
			continue
		}

		driverId, err := strconv.ParseInt(keyParts[1], 10, 64)
		if err != nil {
			continue
		}

		// Get the driver details
		driver, err := GetDriver(driverId)
		if err != nil {
			continue
		}

		drivers = append(drivers, driver)
	}

	return drivers, nil
}
