#!/bin/bash
sleep 10

echo "Initializing replica set..."
mongosh --host mongo1:27017 -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase admin <<EOF
  var cfg = {
    "_id": "rs0",
    "members": [
      { _id: 0, host: "mongo1:27017" },
      { _id: 1, host: "mongo2:27017" },
      { _id: 2, host: "mongo3:27017" }
    ]
  };
  rs.initiate(cfg);
  print("Replica set initialized.");
EOF
echo "Done"