# #https://medium.com/hackernoon/inspecting-docker-images-without-pulling-them-4de53d34a604
get_token() {
  local image=$1

  echo "Retrieving Registry token.
    IMAGE: $image
  " >&2

#   curl \
#     --silent \
#     -u "$DOCKER_USERNAME:$DOCKER_PASSWORD" \
#     "https://auth.docker.io/token?scope=repository:$image:pull&service=registry.docker.io" \
#     | jq -r '.token'
# }

token=GITHUB_PAT
  curl \
    --silent \
    -u "USERNAME:$token" \
    "https://ghcr.io/token?scope=repository:$image:pull&service=ghcr.io" \
    | jq -r '.token'
}

# token=$(get_token merkely/change) 
token=$(get_token merkely-development/merkely-cli)

echo "token:"
echo $token

# curl \
#     --silent -X GET -vvv -k \
#     --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
#     --header "Authorization: Bearer $token" \
#     "https://registry-1.docker.io/v2/merkely/change/manifests/latest" \
#      2>&1 \
#     | grep "< docker-content-digest"

curl \
    --silent -X GET -vvv -k \
    --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    --header "Authorization: Bearer $token" \
    "https://ghcr.io/v2/kosli-dev/merkely-cli/manifests/75405a0" \
     2>&1 \
    | grep "< docker-content-digest"



#https://medium.com/hackernoon/inspecting-docker-images-without-pulling-them-4de53d34a604
get_token() {
  local image=$1

  echo "Retrieving Docker Hub token.
    IMAGE: $image
  " >&2

  curl \
    --silent \
    -u "$REGISTRY_USERNAME:$REGISTRY_PASSWORD" \
    "https://merkelytest.azurecr.io/oauth2/token?service=merkelytest.azurecr.io&scope=repository:simple-server:pull" \
    | jq -r '.access_token'
}

token=$(get_token merkelytest.azurecr.io/simple-server)

curl \
    --silent -X GET -vvv -k \
    --header "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    --header "Authorization: Bearer $token" \
    "https://merkelytest.azurecr.io/v2/simple-server/manifests/v1" \
     2>&1 \
    | grep -i "< docker-content-digest"

