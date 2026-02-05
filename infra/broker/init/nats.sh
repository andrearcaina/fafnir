#!/bin/sh

set -e

echo "Starting NATS setup script..."

NATS_URL="$NATS_HOST://$NATS_HOST:$NATS_PORT"

echo "Attempting to add 'users' stream..."
nats stream add users \
    -s "$NATS_URL" \
    --subjects "users.>" \
    --storage file \
    --retention limits \
    --max-age 7d \
    --max-bytes 100MB \
    --defaults

echo "Successfully added 'users' stream."

echo "Attempting to add 'orders' stream..."
nats stream add orders \
    -s "$NATS_URL" \
    --subjects "orders.>" \
    --storage file \
    --retention limits \
    --max-age 7d \
    --max-bytes 100MB \
    --defaults

echo "Successfully added 'orders' stream."
echo "NATS setup complete."