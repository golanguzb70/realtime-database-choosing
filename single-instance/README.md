# realtime-database-choosing

Short benchmark project to evaluate Redis with RediSearch as a realtime store for driver location and discovery workloads.

## What this benchmark does

- Generates up to 1,000,000 synthetic driver records with realistic fields:
  - `driver_id`, `location` (lat,long), `geo_hash`, `active_tariffs` (TAG), `score` (NUMERIC), `active` (TAG), `phone_charge_percent`, `last_updated_time`.
- Creates a RediSearch index over the `driver:*` hash keys to enable fast geo/text/numeric queries.
- Runs concurrent workload cycles consisting of:
  - Concurrent write updates: upserting drivers continuously.
  - Concurrent single gets: fetching driver by `driver_id`.
  - Concurrent geo-radius list queries: by `@location:[lon lat radius]` sorted by `driver_id`.
  - Concurrent geo+filters list queries: by location, `geo_hash`, `active_tariffs`, `active`, sorted by `score`.

Data generation uses a 2000km area around Tashkent and encodes a precise `geo_hash` for each driver.

## Tech

- Go + `github.com/redis/go-redis/v9`
- Redis with RediSearch (use Redis Stack for convenience)

## Index schema

Created at startup via FT.CREATE on prefix `driver:` with fields:

- `driver_id NUMERIC SORTABLE`
- `location GEO`
- `geo_hash TEXT`
- `active_tariffs TAG SEPARATOR |`
- `score NUMERIC SORTABLE`
- `active TAG`
- `phone_charge_percent NUMERIC NOINDEX`
- `last_updated_time NUMERIC NOINDEX`

## Running the benchmark

1) Start Redis with RediSearch (Redis Stack recommended)

```bash
docker-compose -f redis-image/docker-compose.yml up -d --build
```

2) Clone and build

```bash
git clone https://github.com/golanguzb70/realtime-database-choosing.git
cd realtime-database-choosing
go mod tidy
```

3) Run

```bash
go run .
```

The program will:

- Connect to `localhost:6379`
- Flush the database
- Create the RediSearch index
- Launch concurrent writers/readers for one cycle (default constants in `main.go`)
- Print a summary of write/read ops and errors

## Tuning

Adjust constants in `main.go` to scale load:

- `testCycleCount` – number of one-minute cycles
- `writeGoroutinesCount`, `readGoroutinesCount` – worker counts
- `numDrivers` – ID range for synthetic drivers
- `writeOpsPerMinute`, `singleGetOpsPerMinute`, `multiGetRadOpsPerMinute`, `multiGetGeoHashOpsPerMinute` – target per-minute rates

## Troubleshooting

- If FT.CREATE fails, ensure RediSearch is available (use Redis Stack image).
- High error rates typically indicate rate limits too high for your hardware. Reduce goroutines or ops/min.
- Check Redis logs and system resources if latency spikes.

## Notes

- Synthetic locations are spread within ~2000km of Tashkent and encoded with `github.com/pierrre/geohash`.
- Field names and types in `repository.go` match the FT index schema.