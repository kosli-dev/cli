#!/bin/bash -Eeu

OK=$(echo "rs.status().ok" | \
  mongosh --quiet --norc --username ${MONGO_INITDB_ROOT_USERNAME} \
        --password ${MONGO_INITDB_ROOT_PASSWORD})

test "${OK}" -eq 1
