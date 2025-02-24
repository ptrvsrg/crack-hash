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
  address:
task:
  splitStrategy: chunk-based
```

ENV variables (for example [`config/.env.default`](./config/.env.default)):

```dotenv
SERVER_PORT=8080
SERVER_ENV=dev
WORKER_ADDRESS=
TASK_SPLIT_STRATEGY=chunk-based
```

## Testing

```bash
make test
```