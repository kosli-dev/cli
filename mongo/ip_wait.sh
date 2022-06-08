#!/bin/bash -Eeu

readonly IP_ADDRESS="${1}"
readonly MAX_TRIES=10

# If the server fails to become ready after MAX_TRIES, fail the make target.
# Do not retry endlessly as this can easily burn through hours of CI minutes.

for try in $(seq 1 ${MAX_TRIES}); do
  if [ $(curl -sw '%{http_code}' "${IP_ADDRESS}/ready" -o /dev/null) -eq 200 ]; then
    echo "${IP_ADDRESS} is ready"
    exit 0
  else
    echo "Waiting for ${IP_ADDRESS} readiness... ${try}/${MAX_TRIES}"
    sleep 1
  fi
done
echo "Failed ${IP_ADDRESS} readiness"
docker logs mongo1
exit 1
