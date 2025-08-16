# Smart Meeting Scheduler

A Go-based smart meeting scheduler service that helps find optimal meeting times based on participants' availability and scheduling preferences.

## Architecture

This project follows clean architecture principles and is built using the go-kit framework. The main components are:

- **Domain Layer**: Core business logic and entities
- **Service Layer**: Application business rules and use cases
- **Transport Layer**: HTTP endpoints and request/response handling
- **Repository Layer**: Data persistence and retrieval

### Smart Scheduling Algorithm

The scheduler uses a sophisticated scoring algorithm to find the optimal meeting slot based on several factors:

1. **Time Preference**: Earlier slots during the day are preferred (9 AM - 5 PM)
2. **Calendar Optimization**:
   - Minimizes awkward gaps between meetings
   - Maintains buffer time between meetings (15 minutes)
   - Prefers back-to-back scheduling when appropriate
3. **Working Hours**:
   - Primary hours (9 AM - 5 PM): Highest priority
   - Extended hours (8 AM - 9 AM, 5 PM - 6 PM): Medium priority
   - Off hours: Lowest priority

## Project Structure

```
.
├── cmd/
│   └── server/              # Application entry point
├── internal/
│   ├── domain/             # Business domain models
│   ├── endpoint/           # go-kit endpoints
│   ├── service/            # Business logic implementation
│   └── transport/          # HTTP transport layer
├── pkg/
│   ├── algorithm/          # Scheduling algorithm
│   └── repository/         # Data storage layer
└── scripts/                # Database migrations and utilities
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- MySQL 8.0 or later

### Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up the database:

   ```bash
   # Option 1: Use the setup script (recommended)
   ./scripts/setup_mysql.sh

   # Option 2: Manual setup
   mysql -u root -p -e "CREATE DATABASE meeting_scheduler"
   go run scripts/migrate.go
   ```

4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

**Note**: The application uses environment variables for configuration. You can set them directly or create a `.env` file (or any file you choose):

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=meeting_scheduler

# Server Configuration
PORT=8080
```

See `config.env` for a complete example.

You can specify which environment file to use with the `--env` flag:

```bash
go run cmd/server/main.go --env config.env
```

If not specified, `.env` will be used by default.

### API Endpoints

#### 1. Schedule Meeting

```http
POST /schedule
Content-Type: application/json

{
   "participantIds": ["user1", "user2", "user3"],
   "durationMinutes": 60,
   "timeRange": {
      "start": "2024-09-01T09:00:00Z",
      "end": "2024-09-05T17:00:00Z"
   },
   "title": "Project Meeting" // Optional: If not provided, the meeting will be assigned the default name 'New Meeting'.
}
```

**Note:** The `title` field is optional. If you do not provide a meeting name, the default name "New Meeting" will be assigned.

#### 2. Get User Calendar

```http
GET /users/:userId/calendar?start=2024-09-01T00:00:00Z&end=2024-09-02T00:00:00Z
```

## Testing

Run the tests:

```bash
go test ./...
```

## License

MIT License
