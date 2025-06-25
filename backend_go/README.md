# Language Learning Portal - Backend API

A Go-based backend API for a language learning portal that helps users learn Tamil vocabulary through interactive study sessions.

## Features

- Vocabulary management with Tamil, Romaji, and English translations
- Thematic grouping of words
- Study session tracking
- Progress monitoring and statistics
- RESTful API with JSON responses
- Comprehensive test coverage
- OpenAPI/Swagger documentation

## Tech Stack

- Go 1.21
- Gin Web Framework
- SQLite3 Database
- Testify for testing
- Mage for task automation

## Prerequisites

- Go 1.21 or higher
- SQLite3

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd backend_go
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Initialize the database:
   ```bash
   go run magefiles/main.go initdb
   ```

4. Run migrations:
   ```bash
   go run magefiles/main.go migrate
   ```

5. (Optional) Load sample data:
   ```bash
   go run magefiles/main.go seed
   ```

## Running the Application

Start the server:
```bash
go run main.go
```

The API will be available at `http://localhost:8080/api`

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## API Documentation

The API is documented using OpenAPI/Swagger. You can find the documentation in `swagger.yaml`.

To view the API documentation interactively:
1. Visit [Swagger Editor](https://editor.swagger.io/)
2. Import the `swagger.yaml` file

## API Endpoints

### Dashboard
- GET /api/dashboard/quick-stats
- GET /api/dashboard/last_study_session
- GET /api/dashboard/study_progress

### Words
- GET /api/words
- GET /api/words/:id

### Groups
- GET /api/groups
- GET /api/groups/:id
- GET /api/groups/:id/words
- GET /api/groups/:id/study_sessions

### Study Activities
- GET /api/study_activities/:id
- GET /api/study_activities/:id/study_sessions
- POST /api/study_activities

### Study Sessions
- GET /api/study_sessions
- GET /api/study_sessions/:id
- GET /api/study_sessions/:id/words
- POST /api/study_sessions/:id/words/:word_id/review

### System
- POST /api/reset_history
- POST /api/full_reset

## Project Structure

```
backend_go/
├── api/            # API route definitions
├── internal/
│   ├── handlers/   # Request handlers
│   ├── models/     # Database models
│   └── middleware/ # HTTP middleware
├── migrations/     # Database migrations
├── magefiles/      # Task automation scripts
├── tests/          # Integration tests
├── main.go        # Application entry point
├── swagger.yaml   # API documentation
└── words.db       # SQLite database
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details 