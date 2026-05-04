# GoShort: High-Performance URL Shortener

[Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)
[Redis](https://img.shields.io/badge/Redis-5.0+-DC382D?style=for-the-badge&logo=redis)
[MongoDB](https://img.shields.io/badge/MongoDB-4.4+-47A248?style=for-the-badge&logo=mongodb)



##  Overview

**What is it?**  
GoShort takes long, cumbersome URLs and converts them into compact, easy-to-share short links (e.g., `http://localhost:8080/D7fA9b2E`). 

**What problem does it solve?**  
Long URLs are difficult to read, prone to breaking in emails or text messages, and take up too much space. GoShort solves this by providing reliable short links while tracking how many times they are clicked, all while protecting the system against spam with built-in rate limiting.

**How does it work?**  
At its core, the application generates a unique ID for every new URL using Redis, encodes that ID into a short string using Base62, and stores the mapping in MongoDB. When a user visits the short link, the system fetches the original URL from a high-speed Redis cache and redirects them instantly.

---

## Features

- **Blazing Fast Redirects:** Caches short URL mappings in Redis to minimize database lookups and guarantee sub-millisecond response times.
-  **Persistent Storage:** Safely stores URLs, creation metadata, and click analytics in MongoDB.
-  **Atomic ID Generation:** Utilizes Redis `INCR` to generate guaranteed unique, non-colliding IDs across distributed instances.
- **Base62 Encoding:** Converts numeric IDs to compact, URL-safe 8-character string identifiers.
- **Click Tracking:** Asynchronously tracks the number of times a short URL has been visited without slowing down the redirect response.
-  **Expiration Support (TTL):** Supports optional Time-to-Live, allowing short URLs to automatically expire.
-  **Rate Limiting:** Protects the API against abuse by throttling excessive requests per IP using a Redis-backed Token Bucket algorithm.

---

##  Tech Stack

- **Language:** Go (Golang)
- **Primary Database:** MongoDB (Persistent storage & Analytics)
- **Cache & Rate Limiting:** Redis (High-speed caching & atomic operations)
- **Routing:** Standard `net/http` / Gorilla Mux

---

## Project Architecture

1. **Shorten Request Flow:**
   - The user submits a long URL.
   - The app generates a unique integer using Redis's `INCR` command.
   - This integer is obfuscated and encoded into a Base62 short code.
   - The original URL and short code mapping are saved permanently in MongoDB.
   - The mapping is concurrently cached in Redis for immediate, fast access.

2. **Redirect Request Flow:**
   - The user clicks the short link.
   - The app queries the **Redis cache**. If found *(Cache Hit)*, it immediately redirects the user.
   - If not found *(Cache Miss)*, it queries **MongoDB**. If found, it populates the Redis cache for subsequent requests and redirects the user.
   - A background Go routine asynchronously increments the click counter in MongoDB without blocking the user's redirect.

---

## Setup Instructions

Choose your operating system below to get the project running locally.

### Windows Setup

**1. Prerequisites:**
- Install [Go](https://go.dev/dl/) (1.20 or newer).
- Install [Redis for Windows](https://github.com/microsoftarchive/redis/releases) (Alternatively, use WSL2 or Docker Desktop to run a Redis container).
- Install [MongoDB Community Edition](https://www.mongodb.com/try/download/community) and MongoDB Compass.

**2. Installation:**
Open PowerShell or Command Prompt and run:
```powershell
git clone <your-repository-url>
cd redis
go mod tidy
```

**3. Environment Setup:**
Create a file named `.env` in the root directory:
```env
REDIS_URL=redis://localhost:6379/0
SERVER_PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=url_shortener
```

**4. Running Locally:**
Ensure your Redis and MongoDB services are running in the background. Then start the server:
```powershell
go run cmd/server/main.go
```

### macOS Setup

**1. Prerequisites:**
- Install [Homebrew](https://brew.sh/) if you haven't already.
- Install Go: `brew install go`
- Install Redis: `brew install redis`
- Install MongoDB: `brew tap mongodb/brew` then `brew install mongodb-community@7.0`

**2. Installation:**
Open your Terminal and run:
```bash
git clone <your-repository-url>
cd redis
go mod tidy
```

**3. Environment Setup:**
Create a `.env` file in the root directory:
```env
REDIS_URL=redis://localhost:6379/0
SERVER_PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=url_shortener
```

**4. Running Locally:**
Start your background services using Homebrew:
```bash
brew services start redis
brew services start mongodb-community@7.0
```
Then, start the application server:
```bash
go run cmd/server/main.go
```

---

## Usage & API Endpoints

Once the server is running (default: `http://localhost:8080`), you can test the endpoints using Postman, cURL, or any API client.

### 1. Shorten a URL
Converts a long URL into a short code.

- **Endpoint:** `POST /shorten`
- **Content-Type:** `application/json`

**Request Body:**
```json
{
  "url": "https://www.example.com/some/very/long/path",
  "ttl": 3600 
}
```
*(Note: `ttl` is optional and represents time-to-live in seconds. Omit it for permanent URLs).*

**Response (201 Created):**
```json
{
  "short_url": "http://localhost:8080/D7fA9b2E"
}
```

### 2. Redirect
Accessing the generated short link will redirect you to the original destination.

- **Endpoint:** `GET /{short_code}` (e.g., `GET /D7fA9b2E`)
- **Response:**
  - `302 Found`: Redirects the browser to the original URL.
  - `404 Not Found`: Returned if the short code does not exist or has expired.







