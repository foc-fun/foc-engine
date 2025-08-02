# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FOC Engine is a Starknet-based application with a Go backend (API and indexer) and Cairo smart contracts. The project uses Docker Compose for local development and Helm for Kubernetes deployments.

## Build Commands

### Build Everything
```bash
make build                # Build both engine and contracts
make build-engine         # Build Go binaries only
make build-contracts      # Build Cairo contracts only (requires scarb)
```

### Docker Commands
```bash
make docker-build         # Build Docker images for API and indexer
make docker-push          # Push Docker images to registry
docker compose -f docker-compose-devnet.yml build  # Build local dev environment
```

### Run Commands
```bash
# Using installed CLI
foc-engine run

# Using Docker Compose (local development)
docker compose -f docker-compose-devnet.yml up

# Fresh restart with clean state
docker compose -f docker-compose-devnet.yml down --volumes
docker compose -f docker-compose-devnet.yml build
docker compose -f docker-compose-devnet.yml up
```

### Helm/Kubernetes Commands
```bash
make helm-install         # Deploy to Kubernetes
make helm-upgrade         # Update deployment
make helm-uninstall       # Remove deployment
make helm-template        # Preview generated manifests
```

### Contract Testing
```bash
cd onchain && scarb test  # Run Cairo contract tests using snforge
```

## Architecture

### Backend Services
- **API Service** (`cmd/api/`): REST API server
- **Indexer Service** (`cmd/indexer/`): Starknet event indexer
- Both services use MongoDB and Redis for data storage

### Key Packages
- `internal/config`: Configuration management
- `internal/db/mongo`: MongoDB connection handling
- `internal/provider`: Starknet provider integration
- `internal/registry`: Event processing registry
- `internal/accounts`: Account management

### Smart Contracts
- Located in `onchain/` directory
- Built with Cairo using Scarb
- Test framework: snforge

### Infrastructure
- `infra/`: Helm charts for Kubernetes deployment
- `dockerfiles/`: Docker configurations for each service
- Docker Compose files for local development (devnet) and testnet (sepolia)

## Environment Variables

Required environment variables (see `.env.example`):
- `AVNU_API_KEY`: Required for Helm deployments
- MongoDB and Redis connection settings configured via Docker Compose