# QIS Note Notifier

A Go application that monitors your grade portal (QIS) for changes and notifies you when new grades are posted.

## Features

- 🔄 Automatically checks for grade updates every 30 minutes
- 🔐 Secure credential management via environment variables
- 🐳 Docker support with persistent storage
- 📊 Lightweight and efficient monitoring

## Prerequisites

- Go 1.25+ (for local development)
- Docker & Docker Compose (for containerized deployment)
- Valid QIS credentials

## Quick Start with Docker

1. **Clone and navigate to the project:**
   ```bash
   cd qis-note-notifier
   ```

2. **Set up your credentials:**
   ```bash
   cp .env.example .env
   ```

   Edit `.env` and add your QIS username and password:
   ```
   QIS_USERNAME=your_username
   QIS_PASSWORD=your_password
   ```

3. **Start the application:**
   ```bash
   docker compose up -d
   ```

## Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Set environment variables:**
   ```bash
   export QIS_USERNAME="your_username"
   export QIS_PASSWORD="your_password"
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

## Configuration

The application can be configured using the following environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `QIS_USERNAME` | Your QIS username | - | Yes |
| `QIS_PASSWORD` | Your QIS password | - | Yes |
| `WEBHOOK_URL` | Webhook URL for notifications | - | No |
| `DEBUG` | Enable debug mode (true/false) | false | No |
| `TZ` | Timezone for logs | `Europe/Berlin` | No |

## How It Works

1. The application logs into the QIS portal using your credentials
2. It retrieves your grade overview page
3. A hash of the page content is calculated and compared to the previous hash
4. If changes are detected, you'll see a notification in the logs: 🎉 Änderung erkannt!
5. The process repeats every 30 minutes
