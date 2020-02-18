#!/usr/bin/env bash

docker run \
    --rm \
    --network host \
    -t postman/newman:alpine \
    --env-var host=localhost:3001/api \
    run https://www.getpostman.com/collections/c50f31b0ee2512596fd8 \
    --folder "Endpoints"

docker run \
    --rm \
    --network host \
    -t postman/newman:alpine \
    --env-var host=localhost:3001/api \
    run https://www.getpostman.com/collections/c50f31b0ee2512596fd8 \
    --folder "Seeders"

docker run \
    --rm \
    --network host \
    -t postman/newman:alpine \
    --env-var host=localhost:3001/api \
    run https://www.getpostman.com/collections/c50f31b0ee2512596fd8 \
    --folder "Reports"
