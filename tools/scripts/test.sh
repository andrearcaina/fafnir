#!/bin/bash
set -e

VENV_LOCUST="tests/locust/.venv_locust/bin/locust"

BASE_LOCUST_CMD="$VENV_LOCUST -f tests/locust/locustfile.py --host=http://localhost:8080"

case "$1" in
    locust)
        users="$2"
        spawn_rate="$3"
        run_time="$4"
        headless="$5"

        echo "Starting Locust load test..."

        USER_COUNT="${users:-1000}"
        SPAWN_RATE="${spawn_rate:-10}"
        RUN_TIME="${run_time:-2m}"

        if [ "$headless" = "true" ]; then
            BASE_LOCUST_CMD="$BASE_LOCUST_CMD --headless"
        fi

        $BASE_LOCUST_CMD \
        --users "$USER_COUNT" \
        --spawn-rate "$SPAWN_RATE" \
        --run-time "$RUN_TIME" \
        --html=tests/locust/reports/locust_report.html \
        --csv=tests/locust/reports/locust_report
        ;;
    *)
        echo "Usage: $0 locust_start [users] [spawn_rate] [run_time] [headless]"
        exit 1
    ;;
esac
