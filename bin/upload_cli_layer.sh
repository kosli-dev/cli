# Configuration
TAG=$TAG
AWS_ACCOUNT_ID=$AWS_ACCOUNT_ID
AWS_REGIONS=("eu-central-1" "eu-west-1" "eu-west-2" "eu-west-3" "eu-north-1" "us-east-1" "us-east-2" "us-west-1" "us-west-2") # Where to upload the layer
S3_BUCKET="lambda-layer-mapping-ccc19615fd6c05ace42e71c551995458dbdb1be7"
S3_KEY="lambda_layer_versions.json"
LAYER_NAME="kosli-cli"
RUNTIME="python3.12"
TEMP_FILE="/tmp/lambda_layer_versions.json"
DESCRIPTION="Kosli cli ${TAG}"
ZIP_FILE="lambda_layer.zip"

# Download Kosli cli and pack it to the zip file for the Lambda layer
KOSLI_CLI_VERSION="${TAG:1}"
echo "Downloading Kosli CLI..."
curl -Lo kosli_${KOSLI_CLI_VERSION}_linux_amd64.tar.gz "https://github.com/kosli-dev/cli/releases/download/${TAG}/kosli_${KOSLI_CLI_VERSION}_linux_amd64.tar.gz"
tar -xf kosli_${KOSLI_CLI_VERSION}_linux_amd64.tar.gz -C . && rm kosli_${KOSLI_CLI_VERSION}_linux_amd64.tar.gz
zip -j "$ZIP_FILE" kosli

declare -A REGION_LAYER_MAP

# Iterate through regions
for REGION in "${AWS_REGIONS[@]}"; do
  echo "Publishing Lambda layer to region: $REGION..."
  
  # Publish the Layer
  LAYER_VERSION=$(aws lambda publish-layer-version \
      --region "$REGION" \
      --layer-name "$LAYER_NAME" \
      --description "$DESCRIPTION" \
      --zip-file "fileb://$ZIP_FILE" \
      --compatible-runtimes "$RUNTIME" \
      --query "Version" --output text)
  
  if [ -z "$LAYER_VERSION" ]; then
    echo "Failed to publish the Lambda layer in region $REGION."
    exit 1
  fi
  
  echo "Lambda layer published in region $REGION! Version: $LAYER_VERSION"

  # Make the Layer Publicly Accessible
  echo "Making Lambda layer publicly accessible in region: $REGION..."
  aws lambda add-layer-version-permission \
      --region "$REGION" \
      --layer-name "$LAYER_NAME" \
      --version-number "$LAYER_VERSION" \
      --statement-id "public-access-$REGION" \
      --action "lambda:GetLayerVersion" \
      --principal "*"
  
  echo "Lambda layer is now publicly accessible in region: $REGION!"
  echo "Layer ARN: arn:aws:lambda:${REGION}:${AWS_ACCOUNT_ID}:layer:${LAYER_NAME}:${LAYER_VERSION}"

  # Store the mapping
  REGION_LAYER_MAP["$REGION"]="$LAYER_VERSION"
done

# Update Lambda layer to Kosli cli mapping in the primary region (e.g., first region in the list)
PRIMARY_REGION=${AWS_REGIONS[0]}
echo "Now updating the Lambda layer to Kosli CLI mapping in S3 (Primary region: $PRIMARY_REGION)..."

# Download the existing JSON file from S3
echo "Downloading existing JSON mapping file from S3..."
aws s3 cp "s3://${S3_BUCKET}/${S3_KEY}" "$TEMP_FILE" --region "$PRIMARY_REGION" || {
  echo "Error: Failed to download JSON file from S3. Ensure it exists or check permissions."
  exit 1
}

# Update the JSON file with mappings for all regions
echo "Updating JSON file with new mappings..."
for REGION in "${!REGION_LAYER_MAP[@]}"; do
  LAYER_VERSION=${REGION_LAYER_MAP[$REGION]}
  jq --arg binary "$TAG" --arg region "$REGION" --argjson layer "$LAYER_VERSION" '.[$binary][$region] = $layer' "$TEMP_FILE" > "${TEMP_FILE}.tmp" && mv "${TEMP_FILE}.tmp" "$TEMP_FILE"
done

# Validate the updated JSON file
if ! jq empty "$TEMP_FILE" 2>/dev/null; then
  echo "Error: Updated JSON file is invalid."
  exit 1
fi

# Upload the updated JSON file back to S3
echo "Uploading updated JSON file to S3..."
aws s3 cp "$TEMP_FILE" "s3://${S3_BUCKET}/${S3_KEY}" --region "$PRIMARY_REGION" || {
  echo "Error: Failed to upload updated JSON file to S3."
  exit 1
}

# Clean up
rm -f "$TEMP_FILE"

echo "Update complete. Kosli CLI version $TAG mappings for all regions added to the JSON file."
