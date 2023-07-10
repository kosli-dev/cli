function handler () {
  EVENT_DATA=$1

  S3_OBJECT_KEY=$(echo ${EVENT_DATA} | jq -r ".detail.object.key")

  # Get ECS session id extracting it from the S3 object key
  ECS_EXEC_SESSION_ID="${S3_OBJECT_KEY##*"/"}"
  ECS_EXEC_SESSION_ID="${ECS_EXEC_SESSION_ID%.log*}"
  echo "ECS_EXEC_SESSION_ID is ${ECS_EXEC_SESSION_ID}" 1>&2

  # Download log file from S3 and report it to the Kosli. Use ECS session id as a Kosli workflow id.
  aws s3 cp s3://${LOG_BUCKET_NAME}/${S3_OBJECT_KEY} /tmp/${S3_OBJECT_KEY}

  # Check if workflow already exists. If not - create it.
  KOSLI_WORKFLOWS_LIST=$(./kosli list workflows --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} -o json)
  KOSLI_WORKFLOW_ALREADY_EXISTS=$(echo ${KOSLI_WORKFLOWS_LIST} | jq --arg ECS_EXEC_SESSION_ID "$ECS_EXEC_SESSION_ID" 'any(.id == $ECS_EXEC_SESSION_ID)')

  if [[ $KOSLI_WORKFLOW_ALREADY_EXISTS == false ]]; then
      echo "The Kosli workflow ${ECS_EXEC_SESSION_ID} does not yet exist, creating it..." 1>&2
      ./kosli report workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} --id ${ECS_EXEC_SESSION_ID}
  fi

  # Upload file to the Kosli
  echo "Uploading ECS exec log file to the Kosli..." 1>&2
  ./kosli report evidence workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} \
      -e /tmp/${S3_OBJECT_KEY} --id ${ECS_EXEC_SESSION_ID} \
      --step ${KOSLI_STEP_NAME}
}