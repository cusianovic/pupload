<div align="center">
  <img src="assets/pupload-text.svg" alt="Databasus Logo" width="450"/>
  <h3>Upload service and orchestrator for distributed processing pipelines</h3>
  <p>Pupload is an open source system for handling user upload, applying the media transformations you need, and storing them where you want. Supports writing custom tasks, combining them, and managing the lifetime of results and intermedieries all in one place.</p>

  <!-- <p>
    <a href="#-features">Features</a> •
    <a href="#-installation">Installation</a> •
    <a href="#-usage">Usage</a> •
    <a href="#-license">License</a> •
    <a href="#-contributing">Contributing</a>
  </p> -->
  
  <!-- Badges -->
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/pupload/pupload)](go.mod)
  [![Self Hosted](https://img.shields.io/badge/self--hosted-yes-brightgreen)](https://github.com/pupload/pupload)
  [![Open Source](https://img.shields.io/badge/open%20source-❤️-red)](https://github.com/pupload/pupload)   
</div>


## Quick Start

Requires Go 1.25+, Docker. Optionally uses Redis for stand-alone server.

```bash
# Install
go install github.com/pupload/pupload/cmd/pup@latest

# Download examples
git clone https://github.com/pupload/examples && cd examples

# Initalize
pup init

# Run
pup test image-resize --input image-in=<image_path> --mocks3 
```

 or add your own S3 and kickoff run via CLI or API:

```bash
# Start development server
pup dev

# Test via CLI
pup test image-resize --input image-in=<image_path> --controller localhost:1234

# or API
curl -X POST http://localhost:1234/v1/projects/<project_id>/flows/image-resize
```

## Install

```bash
# Go
go install github.com/pupload/pupload/cmd/pup@latest

# Source
git clone https://github.com/pupload/pupload.git
cd pupload
go build -o pup ./cmd/pup
```

[Releases](https://github.com/pupload/pupload/releases) for binaries.

## Why Pupload?

Most workflow engines are built for generic task orchestration. Pupload is built specifically for files.

**When to use Pupload:**
- You're processing user uploads (images, videos, documents)
- You need different processing tiers (cheap transcoding vs expensive ML models)
- Your pipeline has multiple steps (upload → resize → watermark → store)
- You want to scale workers independently based on GPU/CPU availability

**vs. rolling your own:** You don't have to build presigned URL handling, container orchestration, retry logic, resource scheduling, or distributed state management.

## Use Cases

**Image/Video Processing**
```yaml
upload → thumbnail → watermark → multiple resolutions → CDN
```

**Document Pipeline**
```yaml
PDF upload → extract text → OCR scanned pages → index → store
```

**ML Inference**
```yaml
upload → preprocess → GPU inference → postprocess → results
```

**Archive Processing**
```yaml
zip upload → extract → virus scan → process each file → store
```

Mix CPU and GPU nodes. Scale GPU workers separately. Run expensive steps on dedicated hardware.

## How it works

**Flows** are YAML pipelines. They contain nodes (steps), edges (connections), datawells (input/output), and stores (storage backends).

**Nodes** run in Docker. Each gets presigned URLs for inputs/outputs. On completion, uploads results and triggers downstream nodes.

**Edges** connect nodes in a DAG. Validation prevents cycles.

**Resource tiers**:
- `c-small/medium/large` - CPU
- `m-small/medium/large` - Memory
- `g-small/medium/large` - GPU (NVIDIA, AMD, Intel, Apple)
- `gn-*/ga-*/gi-*` - Vendor-specific

Workers subscribe to resource queues. Controller schedules, workers execute.

## CLI

```bash
pup init                  # New project
pup dev                   # Start controller + worker
pup test <flow>           # Test flow
pup controller list       # List controllers
pup controller add <url>  # Add controller
```

## Configuration

Controller config (`config/controller.yaml`):

```yaml
controller:
  port: 8080
  redis:
    addr: localhost:6379
  storage:
    type: s3
    bucket: uploads
    endpoint: s3.amazonaws.com
```

Worker config (`config/worker.yaml`):

```yaml
worker:
  redis:
    addr: localhost:6379
  docker:
    enabled: true
  resources:
    - c-medium    # Subscribe to medium CPU tasks
    - g-small     # Subscribe to small GPU tasks
```

Workers only pick up tasks matching their subscribed resource tiers. Run cheap workers for CPU tasks, expensive GPU workers only when needed.

## Development

```bash
# Clone
git clone https://github.com/pupload/pupload.git
cd pupload

# Install deps
go mod download

# Run tests
go test ./...

# Run specific package tests
go test ./internal/validation/...

# Build
go build -o pup ./cmd/pup

# Run locally
./pup dev
```


## Contributing

1. Fork the repo
2. Create branch (`git checkout -b feature/thing`)
3. Commit (`git commit -am 'Add thing'`)
4. Push (`git push origin feature/thing`)
5. Open PR
