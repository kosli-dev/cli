#!/bin/bash

force_restart=$1
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
    docker-compose down || true
	docker pull 772819027869.dkr.ecr.eu-central-1.amazonaws.com/merkely:latest || true
	docker-compose up -d
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
docker exec $container_name /demo/create_cli_test_users.py
check_success
