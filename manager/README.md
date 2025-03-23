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
worker:
  addresses:
  health:
    path:
    interval: 1m
    timeout: 1m
    retries: 3
task:
  split:
    strategy: chunk-based
    chunkSize: 10000000
  timeout: 1h
  limit: 10
  maxAge: 24h
  finishDelay: 1m
```

ENV variables (for example [`config/.env.default`](./config/.env.default)):

```dotenv
CONFIG_FILE=config/config.yaml

SERVER_PORT=8080
SERVER_ENV=dev

WORKER_ADDRESSES=
WORKER_HEALTH_PATH=
WORKER_HEALTH_INTERVAL=1m
WORKER_HEALTH_TIMEOUT=1m
WORKER_HEALTH_RETRIES=3

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