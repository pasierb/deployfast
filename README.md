# DeployFast

DeployFast is a Go-based tool designed to simplify and automate the process of deploying applications to remote servers. It handles SSH connections, file transfers, and server provisioning.

## Features

- SSH connection management
- Remote command execution
- File transfer using SFTP
- Automated server provisioning
- Real-time output streaming for provisioning scripts

## Prerequisites

- Go 1.x (version used in your project)
- SSH access to your target server
- Docker (for local development)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/deployfast.git
   cd deployfast
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Build the project:
   ```
   go build -o deployfast cmd/main.go
   ```

## Configuration

Create a `deployfast.json` file in your project root with the following structure:
