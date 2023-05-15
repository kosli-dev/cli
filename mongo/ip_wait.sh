#!/usr/bin/env bash
set -Eeu

readonly URL="${1}"
readonly MAX_TRIES=80

echo -n "Waiting for ${URL} readiness"
for try in $(seq 1 ${MAX_TRIES}); do
  echo -n .
  if [ $(curl -sw '%{http_code}' "${URL}" -o /dev/null) -eq 200 ]; then
    echo
    exit 0
  else
    sleep 0.5
  fi
done

# If the server fails to become ready after MAX_TRIES, fail the make target.
# Do not retry endlessly as this can easily burn through hours of CI minutes.

echo
echo "Failed ${URL} readiness after ${MAX_TRIES} tries"
echo "############### Mongo LOGS ###############"
echo
docker container logs mongo-cli-1
echo 
echo "############### APPLICATION LOGS ###############"
echo
docker container logs cli_kosli_server || true
exit 42
