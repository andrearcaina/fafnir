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
echo "NATS setup complete."