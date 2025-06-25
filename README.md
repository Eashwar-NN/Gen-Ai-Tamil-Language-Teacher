# Language Learning Portal Backend

A Go-based backend API for a language learning portal that manages vocabulary, study sessions, and learning progress.

## Prerequisites

1. Go 1.21 or later
2. GCC (Required for SQLite3)
   - For Windows: Install MinGW-w64 from [https://www.mingw-w64.org/](https://www.mingw-w64.org/)
   - For Linux: `sudo apt-get install gcc`
   - For macOS: Install Xcode Command Line Tools

## Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up the database:
   ```bash
   go run magefiles/magefile.go initdb
   go run magefiles/magefile.go migrate
   ```
4. (Optional) Seed the database with sample data:
   ```bash
   go run magefiles/magefile.go seed
   ```

## Development

### Running the Server

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Running Tests

```bash
go test ./...
```

Note: Tests require CGO and GCC to be available for SQLite3 support.

### API Documentation

The API documentation is available in OpenAPI/Swagger format at `api/swagger.yaml`.

Key endpoints:

- Dashboard: `/api/dashboard/*`
- Words: `/api/words/*`
- Groups: `/api/groups/*`
- Study Sessions: `/api/study_sessions/*`
- System Reset: `/api/reset_history`, `/api/full_reset`

### Project Structure

```
backend_go/
├── api/                    # API documentation
├── internal/
│   ├── database/          # Database initialization and migrations
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # HTTP middleware
│   └── models/            # Data models
├── migrations/            # SQL migration files
├── magefiles/            # Mage task definitions
└── seeds/                # Seed data files
```

## Features

- Complete vocabulary management system
- Study session tracking
- Progress statistics and analytics
- Group-based word organization
- Detailed study history
- System reset capabilities

## Error Handling

The API uses consistent error responses in JSON format:

```json
{
    "errors": [
        {
            "field": "field_name",
            "message": "Error message"
        }
    ]
}
```

## Performance Optimizations

- Database indexes on frequently queried fields
- Pagination support (100 items per page)
- Query optimization for statistics calculations

## Contributing

1. Fork the repository
2. Create your feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License 