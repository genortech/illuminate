# Installation Guide

## Prerequisites

- Go 1.25.1 or later
- Node.js and npm (for frontend assets)
- SQLite (for local development)

## Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/illuminate.git
   cd illuminate
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run the application:
   ```bash
   go run cmd/web/main.go
   ```

## Configuration

Edit the `.env` file to configure your application settings:

```env
APP_ENV=development
PORT=8080
DATABASE_URL=file:illuminate.db
```

## Troubleshooting

If you encounter any issues:

1. Make sure all dependencies are installed
2. Check the logs for error messages
3. Ensure the database is accessible
4. Verify environment variables are set correctly
