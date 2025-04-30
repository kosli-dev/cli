#!/bin/bash
set -euo pipefail

# Validate KOSLI_SERVER_IMAGE
if [[ -z "${KOSLI_SERVER_IMAGE:-}" ]] || [[ "$KOSLI_SERVER_IMAGE" == *"Error"* ]]; then
    echo "❌ Invalid or missing KOSLI_SERVER_IMAGE"
    exit 1
fi

# Set force_restart to the first argument if provided, empty string otherwise
force_restart="${1:-}"
container_name=cli_kosli_server

check_success()
{
    if [ $? -eq 0 ]; then
        echo -e "completed \xE2\x9C\x94"
    else
        echo -e "failed \xE2\x9D\x8C"
        exit 52
    fi
}

restart_server() 
{
    echo restarting server ...
    ./bin/docker_login_aws.sh staging
    docker compose down || true
    echo -e "\033[38;5;208musing server image\033[0m ${KOSLI_SERVER_IMAGE}"
    docker pull ${KOSLI_SERVER_IMAGE} || true
    docker compose up -d
    ./mongo/ip_wait.sh localhost:9010/minio/health/live
    ./mongo/ip_wait.sh localhost:8001/ready
    check_success
}


if [ ! -z "$force_restart" ]; then
    restart_server
elif [ "$( docker container inspect -f '{{.State.Status}}' $container_name )" == "running" ]; then
    echo reseting DB on running server ...
    docker exec $container_name /app/test/clean_database.py 
    check_success
else
    restart_server
fi

echo creating test users on server ...
docker exec $container_name /demo/create_standalone_test_users.py
check_success
