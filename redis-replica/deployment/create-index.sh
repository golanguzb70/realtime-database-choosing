#!/bin/bash
REDIS_HOST=${REDIS_HOST:-127.0.0.1}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_CONNECT_INTERVAL=1   # seconds between PINGs
INDEX_CREATE_INTERVAL=2    # seconds between FT.CREATE retries
INDEX_NAME=index           # RediSearch index name
########################################################

create_index() {
  redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" \
FT.CREATE index ON HASH PREFIX 1 driver: SCHEMA \
driver_id NUMERIC SORTABLE \
location GEO \
geo_hash TEXT \
active_tariffs TAG SEPARATOR \| \
score NUMERIC SORTABLE \
active TAG \
phone_charge_percent NUMERIC NOINDEX \
last_updated_time NUMERIC NOINDEX 
}

index_exists() {
  redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --raw FT._LIST | grep -Fxq "$INDEX_NAME"
}

# 1) Wait until Redis answers PING
# 2) Create the index if it is missing

echo -n "[Redis] Waiting for Redis to accept connections"
until redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" PING &>/dev/null; do
  echo -n "."
  sleep "$REDIS_CONNECT_INTERVAL"
done
echo " up!"

redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" SET test ok >/dev/null

# Create RediSearch index if missing
while true; do
  if index_exists; then
    echo "[Redis] Index '$INDEX_NAME' already exists."
    break
  fi
  echo "[Redis] Attempting to create index '$INDEX_NAME' ..."
  if create_index; then
    echo "[Redis] Index created successfully."
    break
  fi
done

