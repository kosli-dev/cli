function handler () {
  EVENT_DATA=$1

  ECS_EXEC_SESSION_ID=$(echo ${EVENT_DATA} | jq -r ".detail.responseElements.session.sessionId")
  echo "ECS_EXEC_SESSION_ID is ${ECS_EXEC_SESSION_ID}" 1>&2

  # Check if workflow already exists. If not - create it.
  KOSLI_WORKFLOWS_LIST=$(./kosli list workflows --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} -o json)
  KOSLI_WORKFLOW_ALREADY_EXISTS=$(echo ${KOSLI_WORKFLOWS_LIST} | jq --arg ECS_EXEC_SESSION_ID "$ECS_EXEC_SESSION_ID" 'any(.id == $ECS_EXEC_SESSION_ID)')

  if [[ ${KOSLI_WORKFLOW_ALREADY_EXISTS} == false ]]; then
      echo "The Kosli workflow ${ECS_EXEC_SESSION_ID} does not yet exist, creating it..." 1>&2
      ./kosli report workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} --id ${ECS_EXEC_SESSION_ID}
  fi

  # Get ECS exec session user (ARN of the IAM role that initiated the session)
  ECS_EXEC_USER=$(echo ${EVENT_DATA} | jq -r ".detail.userIdentity.arn")
  echo "{\"ecs_exec_role_arn\": \"${ECS_EXEC_USER}\"}" | jq . > /tmp/user-identity.json

  echo "Reporting ECS exec user data to the Kosli..." 1>&2
  ./kosli report evidence workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} \
      --user-data /tmp/user-identity.json \
      --evidence-paths /tmp/user-identity.json \
      --id ${ECS_EXEC_SESSION_ID} \
      --step ${KOSLI_STEP_NAME}
}
