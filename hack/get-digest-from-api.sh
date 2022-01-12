#https://medium.com/hackernoon/inspecting-docker-images-without-pulling-them-4de53d34a604
get_token() {
  local image=$1

  echo "Retrieving Docker Hub token.
    IMAGE: $image
  " >&2

  curl \
    --silent \
    -u "$DOCKER_USERNAME:$DOCKER_PASSWORD" \
    "https://auth.docker.io/token?scope=repository:$image:pull&service=registry.docker.io" \
    | jq -r '.token'
}

token=$(get_token merkely/change)

curl \
    --silent -X GET -vvv -k \
    --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    --header "Authorization: Bearer $token" \
    "https://registry-1.docker.io/v2/merkely/change/manifests/latest" \
     2>&1 \
    | grep "< docker-content-digest"

