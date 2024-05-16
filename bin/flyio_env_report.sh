#!/usr/bin/env bash

SCRIPT_NAME=flyio_env_report.sh

USAGE="$SCRIPT_NAME [options] <protocol://hostname> <owner> <environment-name>"




die()
{
    echo "Error: $1" >&2
    exit 1
}

print_help()
{
    cat <<EOF
Usage: "$USAGE"
Options are:
  -h          Print this help message
EOF
}

check_arguments()
{
    while getopts "h" opt; do
      case $opt in
        h)
        print_help
        exit 1
        ;;
        \?)
        echo "Invalid option: -$OPTARG" >&2
        exit 1
        ;;
      esac
    done

    if [ ! $# -eq 3 ]; then
        die "Not enough arguments: $USAGE"
    fi

    HOST=$1
    OWNER=$2
    ENV_NAME=$3
}

loud_curl()
{
  # curl that prints the server traceback if the response
  # status code is not in the range 200-299
  local -r TYPE="${1}"
  local -r URL="${2}"
  local -r JSON_PAYLOAD="${3}"
  local -r OUTPUT_FILE=$(mktemp)
  set +e
  HTTP_CODE=$(curl --header 'Content-Type: application/json' \
       --user "${KOSLI_API_TOKEN}":unused \
       --output "${OUTPUT_FILE}" \
       --write-out "%{http_code}" \
       --request "${TYPE}" \
       --silent \
       --data "${JSON_PAYLOAD}" \
       "${URL}")
  set -e
  >&2 cat "${OUTPUT_FILE}"
  if [[ ${HTTP_CODE} -lt 200 || ${HTTP_CODE} -gt 299 ]] ; then
      echo "Output at ${OUTPUT_FILE}"
      exit 22
  fi
  rm "${OUTPUT_FILE}"
}

 

main()
{
    check_arguments "$@"
    local image=$(flyctl image show --json)
    local name=$(echo "${image}" | jq .Repository | sed 's/"//g')
    local tag=$(echo "${image}" | jq .Tag | sed 's/"//g')
    local fingerprint=$(echo "${image}" | jq .Digest | sed "s/sha256://" | sed 's/"//g')

    local status=$(flyctl status --json)
    local createdAtStr=$(echo "${status}" | jq .Allocations[0].CreatedAt | sed 's/"//g')

    if [ $(uname) = "Linux" ]; then
        local createdAt=$(date --date="${createdAtStr}" +%s)
    else
        local createdAt=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "${createdAtStr}" +%s)
    fi

    local json_data=$( jq -n \
                  --arg nameTag "${name}:${tag}" \
                  --arg fp "$fingerprint" \
                  --argjson ts "$createdAt" \
                  --arg envName "$ENV_NAME" \
                  '{artifacts: [{digests: {($nameTag): $fp}, creationTimestamp: $ts}], type: "server", id: $envName}')
    loud_curl PUT "${HOST}/api/v1/environments/${OWNER}/${ENV_NAME}/data" "${json_data}"
}

main "$@"
