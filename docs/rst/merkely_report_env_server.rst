.. list-table:: merkely report env server
   :header-rows: 1

   * - ENV_VAR_NAME
     - Required?
     - Notes
   * - MERKELY_API_TOKEN
     - yes
     - the merkely API token.
   * - MERKELY_CONFIG_FILE
     - no
     - [optional] the merkely config file path. Defaults to :code:`merkely`.
   * - MERKELY_DRY_RUN
     - no
     - whether to send the request to the endpoint or just log it in stdout. Defaults to :code:`false`.
   * - MERKELY_HOST
     - no
     - the merkely endpoint. Defaults to :code:`https://app.merkely.com`.
   * - MERKELY_ID
     - yes
     - the unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.
   * - MERKELY_MAX_API_RETRIES
     - no
     - how many times should API calls be retried when the API host is not reachable. Defaults to :code:`3`.
   * - MERKELY_OWNER
     - yes
     - the merkely organization.
   * - MERKELY_PATHS
     - no
     - the comma separated list of artifact directories. Defaults to :code:`[]`.
