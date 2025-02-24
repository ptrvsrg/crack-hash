<h1 align="center">Crack Hash</h1>

<p align="center">
  <img alt="License" src="https://img.shields.io/github/license/ptrvsrg/crack-hash?color=56BEB8&style=flat">
  <img alt="Github issues" src="https://img.shields.io/github/issues/ptrvsrg/crack-hash?color=56BEB8&style=flat" />
  <img alt="Github forks" src="https://img.shields.io/github/forks/ptrvsrg/crack-hash?color=56BEB8&style=flat" />
  <img alt="Github stars" src="https://img.shields.io/github/stars/ptrvsrg/crack-hash?color=56BEB8&style=flat" />
</p>

Distributed system for cracking MD5 hashes

## Quickstart

### Requirements

- Golang 1.23.5
- Docker
- Docker Compose
- Make (optional)

### Clone repository

```bash
git clone https://github.com/ptrvsrg/crack-hash.git
```

### Running

#### Docker

```bash
docker compose up -d --force-recreate --build
```

#### Manually

+ Build manager and worker:

```bash
make -f ./manager/Makefile -C ./manager build
```

```bash
make -f ./worker/Makefile -C ./worker build
```

+ Run manager and worker:

```bash
./manager/bin/manager server
```

```bash
./manager/bin/worker server
```

## License

This project is distributed under the [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0.html) license