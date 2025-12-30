# Food Sharing Backend API

A RESTful API backend for a food sharing application built with Go, Echo framework, PostgreSQL, and MinIO.

## Features

- **User Authentication**: JWT-based authentication with bcrypt password hashing
- **Food Posts**: Create, read, update, and delete food posts with image upload support
- **Hunger Broadcasts**: Allow users to broadcast their need for food
- **Request Management**: Users can request food and owners can accept/reject requests
- **Feed System**: Combined feed of food posts and hunger broadcasts with filtering
- **Image Storage**: MinIO integration for secure image storage
- **Database Migrations**: Versioned database schema management

## Tech Stack

- **Framework**: Echo v4
- **Database**: PostgreSQL
- **Object Storage**: MinIO
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt
- **Validation**: go-playground/validator
- **Migration**: golang-migrate

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- MinIO server
- golang-migrate CLI (for migrations)

## Installation

### 1. Clone the repository

```bash
cd c:\workspace\pocs\starving\v0-hungry-food-sharing-app-backend
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Install golang-migrate CLI

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 4. Configure environment variables

Copy the example environment file and update with your credentials:

```bash
cp .env.example .env
```

Edit `.env` with your database and MinIO credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=food_sharing
DB_SSLMODE=disable

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=food-images
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000

JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION=24h

PORT=8000
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

MAX_UPLOAD_SIZE=5242880
```

### 5. Run database migrations

```bash
migrate -path migrations -database "postgres://postgres:yourpassword@localhost:5432/food_sharing?sslmode=disable" up
```

Or use the Makefile (update DB credentials in Makefile first):

```bash
make migrate-up
```

### 6. Start the server

```bash
go run main.go
```

Or use the Makefile:

```bash
make run
```

The server will start on `http://localhost:8000`

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and receive JWT token
- `GET /api/auth/me` - Get current user (requires authentication)

### Feed

- `GET /api/feed` - Get combined feed (supports `?type=all|food|hunger`)

### Food Posts

- `POST /api/food-posts` - Create food post (with optional image upload)
- `GET /api/food-posts` - Get all food posts
- `GET /api/food-posts/:id` - Get specific food post
- `PUT /api/food-posts/:id` - Update food post (owner only)
- `DELETE /api/food-posts/:id` - Delete food post (owner only)
- `POST /api/food-posts/:id/request` - Request food from post
- `GET /api/food-posts/:id/requests` - Get all requests for post (owner only)
- `PUT /api/food-posts/:postId/requests/:requestId/accept` - Accept request (owner only)
- `PUT /api/food-posts/:postId/requests/:requestId/reject` - Reject request (owner only)

### Hunger Broadcasts

- `POST /api/hunger-broadcasts` - Create hunger broadcast
- `GET /api/hunger-broadcasts` - Get all broadcasts
- `GET /api/hunger-broadcasts/:id` - Get specific broadcast
- `DELETE /api/hunger-broadcasts/:id` - Delete broadcast (owner only)
- `POST /api/hunger-broadcasts/:id/offer` - Offer food to broadcast

### User

- `GET /api/my-requests` - Get all requests made by current user
- `GET /api/my-posts` - Get all posts by current user
- `GET /api/my-hunger-broadcasts` - Get all broadcasts by current user
- `GET /api/profile` - Get user profile with statistics
- `PUT /api/profile` - Update user profile

### Health Check

- `GET /health` - Server health check

## Project Structure

```
.
├── config/              # Configuration and initialization
│   ├── config.go        # Environment configuration
│   ├── database.go      # Database connection
│   └── minio.go         # MinIO client setup
├── handlers/            # HTTP request handlers
│   ├── auth_handler.go
│   ├── feed_handler.go
│   ├── food_post_handler.go
│   ├── hunger_broadcast_handler.go
│   └── user_handler.go
├── middleware/          # Custom middleware
│   ├── auth.go          # JWT authentication
│   └── cors.go          # CORS configuration
├── migrations/          # Database migrations
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   └── ...
├── models/              # Data models
│   ├── user.go
│   ├── food_post.go
│   ├── hunger_broadcast.go
│   ├── food_request.go
│   └── hunger_offer.go
├── repository/          # Database operations
│   ├── user_repository.go
│   ├── food_post_repository.go
│   ├── hunger_broadcast_repository.go
│   ├── food_request_repository.go
│   └── hunger_offer_repository.go
├── routes/              # Route definitions
│   └── routes.go
├── services/            # Business logic
│   ├── auth_service.go
│   ├── feed_service.go
│   ├── food_post_service.go
│   ├── hunger_broadcast_service.go
│   └── image_service.go
├── utils/               # Utility functions
│   ├── jwt.go
│   ├── password.go
│   ├── response.go
│   └── validator.go
├── .env.example         # Environment variables template
├── .gitignore
├── go.mod
├── main.go              # Application entry point
├── Makefile             # Build and run commands
└── README.md
```

## Development

### Run the server

```bash
make run
```

### Build the binary

```bash
make build
```

### Run tests

```bash
make test
```

### Create a new migration

```bash
make migrate-create
# Enter migration name when prompted
```

### Rollback migrations

```bash
make migrate-down
```

## Testing with cURL

### Register a user

```bash
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'
```

### Login

```bash
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### Create a food post (replace TOKEN with your JWT)

```bash
curl -X POST http://localhost:8000/api/food-posts \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Fresh Pizza",
    "description":"2 large pizzas left from party",
    "quantity":"2 pizzas",
    "location":"Downtown",
    "expiryDate":"2025-01-01T20:00:00Z"
  }'
```

### Get feed

```bash
curl http://localhost:8000/api/feed
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | - |
| `DB_NAME` | Database name | `food_sharing` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` |
| `MINIO_BUCKET` | MinIO bucket name | `food-images` |
| `MINIO_USE_SSL` | Use SSL for MinIO | `false` |
| `MINIO_PUBLIC_URL` | Public URL for MinIO | `http://localhost:9000` |
| `JWT_SECRET` | JWT signing secret | - |
| `JWT_EXPIRATION` | JWT token expiration | `24h` |
| `PORT` | Server port | `8000` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |
| `MAX_UPLOAD_SIZE` | Max file upload size (bytes) | `5242880` (5MB) |

## License

MIT
