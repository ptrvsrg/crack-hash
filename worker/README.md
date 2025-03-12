<h1 align="center">Crack Hash - Worker</h1>

Worker service receives request from manager and executes it.

## CLI help

```
NAME:
   ./bin/worker - the cli application for Crack-Hash worker

USAGE:
   ./bin/worker [global options] [command [command options]]

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
manager:
  address:
task:
  split:
    strategy: chunk-based
    chunk-size: 10000000
  concurrency: 1000
```

ENV variables (for example [`config/.env.default`](./config/.env.default)):

```dotenv
CONFIG_FILE=config/config.yaml

SERVER_PORT=8080
SERVER_ENV=dev

MANAGER_ADDRESS=

TASK_SPLIT_STRATEGY=chunk-based
TASK_SPLIT_CHUNK_SIZE=10000000
TASK_CONCURRENCY=1000
```

## Makefile

```bash
Available commands:
  build                 - Build the application
  build-image   - Build the docker image
  run                   - Run the application (set the COMMAND environment variable to change the command, default is 'server')
  swagger               - Generate Swagger specification
  mock                  - Generate mocks
  lint                  - Lint the application
  test                  - Test the application
  clean                 - Clean the binary
  watch                 - Live Reload
```