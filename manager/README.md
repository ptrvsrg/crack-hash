<h1 align="center">Crack Hash - Manager</h1>

Manager service receives request from user and distributes it to workers.

## CLI help

```
NAME:
   ./bin/manager - The cli application for Crack-Hash manager

USAGE:
   ./bin/manager [global options] [command [command options]]

VERSION:
   0.0.0-local

AUTHOR:
   ptrvsrg

COMMANDS:
   server, s       Start the server
   healthcheck, H  Healthcheck
   version, v      Print the Version
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

COPYRIGHT:
   © 2025 ptrvsrg
```

## Configuration

YAML file (for example [`config/config.default.yaml`](./config/config.default.yaml)):

```yaml
server:
  port: 8080
  env: dev
  cors:
    allowedOrigins:
      - "*"
    allowedMethods:
      - GET
      - POST
      - PUT
      - PATCH
      - DELETE
      - OPTIONS
    allowedHeaders:
      - "*"
    allowCredentials: false
    maxAge: 24h
mongodb:
  uri:
  username:
  password:
  db:
  writeconcern:
    w: majority
    journal:
  readconcern:
    level: majority
amqp:
  uris:
  username:
  password:
  prefetch: 20
  consumers:
    taskresult:
      queue:
  publishers:
    taskstarted:
      exchange:
      routingkey:
task:
  alphabet: abcdefghijklmnopqrstuvwxyz0123456789
  split:
    strategy: chunk-based
    chunksize: 10000000
  timeout: 1h
  limit: 10
  maxage: 24h
  restartdelay: 1m
  finishdelay: 1m
```

ENV variables (for example [`config/.env.default`](./config/.env.default)):

```dotenv
CONFIG_FILE=config/config.yaml

SERVER_PORT=8080
SERVER_ENV=dev
SERVER_CORS_ALLOWEDORIGINS=*
SERVER_CORS_ALLOWEDMETHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
SERVER_CORS_ALLOWEDHEADERS=*
SERVER_CORS_ALLOWCREDENTIALS=false
SERVER_CORS_MAXAGE=24h

MONGODB_URI=
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_DB=
MONGODB_WRITECONCERN_W=majority
MONGODB_WRITECONCERN_JOURNAL=
MONGODB_READCONCERN_LEVEL=majority

AMQP_URIS=
AMQP_USERNAME=
AMQP_PASSWORD=
AMQP_PREFETCH=20

AMQP_CONSUMERS_TASKRESULT_QUEUE=

AMQP_PUBLISHERS_TASKSTARTED_EXCHANGE=
AMQP_PUBLISHERS_TASKSTARTED_ROUTINGKEY=

TASK_ALPHABET=abcdefghijklmnopqrstuvwxyz0123456789
TASK_SPLIT_STRATEGY=chunk-based
TASK_SPLIT_CHUNK_SIZE=10000000
TASK_TIMEOUT=1h
TASK_LIMIT=10
TASK_MAX_AGE=24h
TASK_FINISH_DELAY=1m
```

## Makefile

```bash
Available commands:
  build                 - Build the application
  build-image           - Build the docker image
  run                   - Run the application (set the COMMAND environment variable to change the command, default is 'server')
  swagger               - Generate Swagger specification
  mock                  - Generate mocks
  lint                  - Lint the application
  test                  - Test the application
  clean                 - Clean the binary
  watch                 - Live Reload
```