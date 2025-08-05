# `ping`

A simple chat API server written in Go.
It is designed to be a lightweight and easy-to-use solution for building chat applications.

## Features
- **WebSocket Support**: Real-time communication using WebSockets.
- **REST API**: Provides a RESTful interface for chat operations.
- **SQLite Database**: Uses SQLite for data storage, making it easy to set up and manage.
- **Authentication**: Supports user authentication and authorization.

## Prerequisites
- Go 1.20 or later
- SQLite3 installed on your system
- Redis

## Installation
1. Clone the repository:
```bash
git clone https://github.com/rtk-rnjn/ping.git
cd ping
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o ping
```

4. Run the application:
```bash
./ping
```

## Configuration
The application can be configured using environment variables. Rename `example.env` to `.env`. The following variables are available:
- `JWT_SECRET`: Secret key used for JWT signing and verification.
- `REDIS_ADDR`: Address of the Redis server (if using Redis for session management).
- `HOST`: Host address to bind the server (default is `0.0.0.0`).
- `PORT`: Port number to bind the server (default is `8080`).

## Usage
Once the application is running, you can access the API at `http://127.0.0.1:8080`.
