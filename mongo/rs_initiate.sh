#!/bin/bash -Eeu

# Call rs.initiate(config) on one member of the replica-set.
# The member receiving the configuration passes it to the other members.

date +"%T"
echo "Starting replica set initialize"
for n in $(seq 1); do
  until mongo --host "mongo${n}" --eval "print(\"waited for connection\")"; do
      echo -n .; sleep 2
  done
done
echo "Connection finished"


MONGO1_IP=$(getent hosts mongo1 | awk '{ print $1 }')

echo "Creating replica set"
mongo --host mongo1 \
      --username "${MONGO_INITDB_ROOT_USERNAME}" \
      --password "${MONGO_INITDB_ROOT_PASSWORD}" \
<<EOF
const config = {
    _id : "rs0",
    version: 1,
    members: [
        {
            "_id": 1,
            "host": "$MONGO1_IP:27017",
            "priority": 3
        }
    ]
};
rs.initiate(config, { force: true });
EOF
echo "replica set created"
