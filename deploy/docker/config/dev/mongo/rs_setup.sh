#!/bin/bash
sleep 10

echo "Initializing replica set..."
mongosh --host mongo1:27017 -u $MONGO_USERNAME -p $MONGO_PASSWORD --authenticationDatabase admin <<EOF
  var cfg = {
    "_id": "rs0",
    "members": [
      { _id: 0, host: "172.28.0.10:27017" },
      { _id: 1, host: "172.28.0.11:27017" },
      { _id: 2, host: "172.28.0.12:27017" }
    ]
  };
  rs.initiate(cfg);
  print("Replica set initialized.");
EOF
echo "Done"