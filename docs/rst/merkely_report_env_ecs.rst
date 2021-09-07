.. list-table:: merkely report env ecs
   :header-rows: 1

   * - ENV_VAR_NAME
     - Default
     - Notes
   * - MERKELY_API_TOKEN
     - 
     - the merkely API token.
   * - MERKELY_CLUSTER
     - 
     - name of the ECS cluster
   * - MERKELY_CONFIG_FILE
     - merkely
     - [optional] the merkely config file path.
   * - MERKELY_DRY_RUN
     - false
     - whether to send the request to the endpoint or just log it in stdout.
   * - MERKELY_HOST
     - https://app.merkely.com
     - the merkely endpoint.
   * - MERKELY_MAX_API_RETRIES
     - 3
     - how many times should API calls be retried when the API host is not reachable.
   * - MERKELY_OWNER
     - 
     - the merkely organization.
   * - MERKELY_SERVICE_NAME
     - 
     - name of the ECS service
