# High-Performance Go URL Shortener

A production-ready URL shortener service built with **Go**, **MongoDB** (for persistent storage and analytics), and **Redis** (for high-speed caching and atomic ID generation).

## Features

- **Blazing Fast Redirects**: Caches short URL mappings in Redis to minimize database lookups.
- **Persistent Storage**: Safely stores URLs, metadata, and click analytics in MongoDB.
- **Atomic ID Generation**: Uses Redis `INCR` to generate guaranteed unique, non-colliding IDs.
- **Base62 Encoding**: Converts numeric IDs to compact, URL-safe 8-character string identifiers.
- **Click Tracking**: Asynchronously tracks the number of times a short URL has been visited without slowing down the redirect response.
- **Expiration Support (TTL)**: Supports optional Time-to-Live for short URLs.

## Architecture

1. **Shorten Request**:
   - The application generates a unique integer using Redis's `INCR` command.
   - This integer is obfuscated and encoded into a Base62 short code.
   - The original URL and short code mapping are saved permanently in MongoDB.
   - The mapping is concurrently cached in Redis for fast access.
2. **Redirect Request**:
   - The application first queries Redis. If the URL is found (Cache Hit), it immediately redirects the user.
   - If the URL is not in Redis (Cache Miss), it falls back to MongoDB. If found, it populates the Redis cache for subsequent requests and redirects the user.
   - In both cases, a background Go routine asynchronously increments the click counter in MongoDB.

## Prerequisites

To run this project locally, you will need:
- [Go](https://golang.org/doc/install) (1.20 or later)
- [Redis](https://redis.io/download) server running locally or remotely.
- [MongoDB](https://www.mongodb.com/try/download/community) server running locally or remotely.

## Setup & Installation

1. **Navigate to the project directory** (assuming you already have the files):
   ```bash
   cd redis
   ```

2. **Install Dependencies**:
   ```bash
   go get go.mongodb.org/mongo-driver/mongo
   go mod tidy
   ```

3. **Environment Configuration**:
   The application uses a `.env` file at the root of the project. A typical `.env` looks like this:
   ```env
   REDIS_URL=redis://localhost:6379/0
   SERVER_PORT=8080
   MONGO_URI=mongodb://localhost:27017
   MONGO_DB_NAME=url_shortener
   ```

## Running the Application

Start the server using:
```bash
go run cmd/server/main.go
```
You should see terminal logs indicating successful connections to both Redis and MongoDB.

## API Documentation

### 1. Shorten a URL

**Endpoint:** `POST /shorten`
**Content-Type:** `application/json`

**Request Body:**
```json
{
  "url": "https://www.example.com/some/very/long/path",
  "ttl": 3600 
}
```
*(Note: `ttl` is optional and represents seconds. If omitted or set to 0, the URL never expires.)*

**Response (201 Created):**
```json
{
  "short_url": "http://localhost:8080/D7fA9b2E"
}
```

### 2. Redirect

**Endpoint:** `GET /{short_code}`

**Response:**
- `302 Found`: Redirects the browser to the original URL.
- `404 Not Found`: If the short code does not exist or has expired.

## Project Structure

```text
.
├── .env                        # Environment variables
├── cmd/
│   └── server/
│       └── main.go             # Application entry point & service wiring
├── internal/
│   ├── config/
│   │   └── config.go           # Environment variable parsing
│   ├── handlers/
│   │   └── url_handler.go      # HTTP handlers for API endpoints
│   ├── models/
│   │   └── url.go              # Data structures and BSON/JSON tags
│   ├── services/
│   │   └── url_service.go      # Core business logic (DB & Cache interactions)
│   └── utils/
│       └── base62.go           # ID hashing and Base62 encoding logic
├── go.mod                      # Go module dependencies
└── README.md                   # Project documentation
```
