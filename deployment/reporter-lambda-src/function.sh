function handler () {
  EVENT_DATA=$1
  echo "$EVENT_DATA" 1>&2;
  ./kosli snapshot ecs $KOSLI_ENV -C $ECS_CLUSTER --org $KOSLI_ORG
}