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

  # Get ECS exec session user identity (ARN of the IAM role that initiated the session)
  ECS_EXEC_USER_IDENTITY=$(echo ${EVENT_DATA} | jq -r ".detail.userIdentity.arn")
  echo "{\"ecs_exec_role_arn\": \"${ECS_EXEC_USER_IDENTITY}\"}" | jq . > /tmp/user-identity.json

  echo "Reporting ECS exec user identity to the Kosli..." 1>&2
  ./kosli report evidence workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} \
      --user-data /tmp/user-identity.json \
      --evidence-paths /tmp/user-identity.json \
      --id ${ECS_EXEC_SESSION_ID} \
      --step ${KOSLI_STEP_NAME_USER_IDENTITY}

  # Get ECS exec session service identity
  echo "Getting ECS task ARN..." 1>&2
  ECS_EXEC_TASK_ARN=$(echo ${EVENT_DATA} | jq -r ".detail.responseElements.taskArn")
  echo "ECS task ARN is ${ECS_EXEC_TASK_ARN}" 1>&2
  echo "Getting ECS Cluster name..." 1>&2
  ECS_EXEC_CLUSTER=$(echo ${EVENT_DATA} | jq -r ".detail.requestParameters.cluster")
  echo "ECS Cluster name is ${ECS_EXEC_CLUSTER}" 1>&2
  echo "Getting ECS task group..." 1>&2
  ECS_EXEC_TASK_GROUP=$(aws ecs describe-tasks --cluster ${ECS_EXEC_CLUSTER} --tasks ${ECS_EXEC_TASK_ARN} | jq ".tasks[].group")
  echo "ECS task group is ${ECS_EXEC_TASK_GROUP}" 1>&2

  echo "{\"ecs_exec_service_identity\": ${ECS_EXEC_TASK_GROUP}}" | jq . > /tmp/service-identity.json

  echo "Reporting ECS exec service identity to the Kosli..." 1>&2
  ./kosli report evidence workflow --audit-trail ${KOSLI_AUDIT_TRAIL_NAME} \
      --user-data /tmp/service-identity.json \
      --evidence-paths /tmp/service-identity.json \
      --id ${ECS_EXEC_SESSION_ID} \
      --step ${KOSLI_STEP_NAME_SERVICE_IDENTITY}
}