.. list-table:: merkely report env k8s
   :header-rows: 1

   * - ENV_VAR_NAME
     - Default
     - Notes
   * - MERKELY_API_TOKEN
     - 
     - the merkely API token.
   * - MERKELY_CONFIG_FILE
     - merkely
     - [optional] the merkely config file path.
   * - MERKELY_DRY_RUN
     - false
     - whether to send the request to the endpoint or just log it in stdout.
   * - MERKELY_EXCLUDE_NAMESPACE
     - []
     - the comma separated list of namespaces (or namespaces regex patterns) NOT to harvest artifacts info from. Can't be used together with --namespace.
   * - MERKELY_HOST
     - https://app.merkely.com
     - the merkely endpoint.
   * - MERKELY_KUBECONFIG
     - 
     - kubeconfig path for the target cluster
   * - MERKELY_MAX_API_RETRIES
     - 3
     - how many times should API calls be retried when the API host is not reachable.
   * - MERKELY_NAMESPACE
     - []
     - the comma separated list of namespaces (or namespaces regex patterns) to harvest artifacts info from. Can't be used together with --exclude-namespace.
   * - MERKELY_OWNER
     - 
     - the merkely organization.
