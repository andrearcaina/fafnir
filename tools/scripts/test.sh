#!/bin/bash

set -e

BASE_LOCUST_CMD="locust -f tests/locust/locustfile.py --host=http://localhost:8080"

case "$1" in
    locust_start)
        echo "Starting Locust load test..."

        if [ -z "$users" ]; then
            echo "No user count specified. Using default of 1000 users."
            USER_COUNT=1000
        else
            USER_COUNT=$users
        fi

        if [ -z "$spawn_rate" ]; then
            echo "No spawn rate specified. Using default of 10 user per second."
            SPAWN_RATE=10
        else
            SPAWN_RATE=$spawn_rate
        fi

        if [ "$headless" == "true" ]; then
            BASE_LOCUST_CMD="$BASE_LOCUST_CMD --headless"
        fi

        if [ -z "$run_time" ]; then
            echo "No run time specified. Using default of 2 minutes."
            run_time="2m"
        else
            BASE_LOCUST_CMD="$BASE_LOCUST_CMD --run-time $run_time"
        fi

        $BASE_LOCUST_CMD --users $USER_COUNT --spawn-rate $SPAWN_RATE --run-time $run_time --html=tests/locust/reports/locust_report.html --csv=tests/locust/reports/locust_report
        ;;
    *)
        echo "Usage: $0 {locust_start}"
        exit 1
        ;;
esac
