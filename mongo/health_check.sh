#!/bin/bash -Eeu

OK=$(echo "rs.status().ok" | \
  mongo --username ${MONGO_INITDB_ROOT_USERNAME} \
        --password ${MONGO_INITDB_ROOT_PASSWORD} \
        --quiet)

test "${OK}" -eq 1
