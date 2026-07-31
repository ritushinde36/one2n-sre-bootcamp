# Student REST API

A simple Go-based REST API for doing basic CRUD operations on student records using Gin Gonic, GORM, and MySQL.

## Features

- Create, read, update, and delete student records
- MySQL database integration with go using GORM
- Router set up using Gin and Gonic.
- Versioned database schema migrations via goose
- API served on port 8888

## Project Structure

- main.go - application entry point. Connects to DB and sets up routes.
- controllers/StudentController.go - request handlers for CRUD operations
- models/studentModel.go - Student data model
- connections/db_connection.go - MySQL connection setup
- config/load_congig.go - for loading the db creds into the application
- migrations/ - goose SQL migration files for the database schema
- cmd/migrate - standalone command to run migrations (up/down/status)
- controllers_test - testing the methods

## Prerequisites

Before running the application, make sure you have:

- Go installed
- MySQL running locally
- A database named student_db

## Local Setup (without Docker)

1. Create the database in MySQL:

```sql
CREATE DATABASE student_db;
```

2. Install Go dependencies:

```bash
go mod tidy
```

3. Create a `.env` file in the project root with your MySQL DSN:

```bash
DSN=root:yourpassword@tcp(127.0.0.1:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
```

4. Run database migrations to create the schema:

```bash
make migrate-up
```

5. Run the server:

```bash
make run
```

The API will start on:

```text
http://localhost:8888
```

## Running with Docker

This runs the API and MySQL as separate containers on a shared Docker
network, instead of relying on a locally installed MySQL.

### Prerequisites

- Docker installed and running

### 1. Create a `.env.docker` file

This is separate from the `.env` used above - it's specifically for the
containerized setup, since the DSN needs to point at the MySQL *container*
(`student-mysql`) rather than `127.0.0.1`. It's gitignored, same as `.env`.

```bash
MYSQL_ROOT_PASSWORD=yourpassword
MYSQL_DATABASE=student_db
DSN=root:yourpassword@tcp(student-mysql:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
PORT=8888
```

### 2. Start MySQL as a container

```bash
make docker-mysql-up
```

This creates a Docker network (`student-api-net`), and starts a MySQL
container on it with a persistent volume, so data survives even if the
container is removed and recreated later. If you have a local MySQL
already running on port 3306 (e.g. via Homebrew), stop it first
(`brew services stop mysql`) to free up the port.

### 3. Run migrations against the containerized MySQL

```bash
make migrate-up
```

This runs directly on your machine (not in a container), reaching the
MySQL container through the port published to your host - make sure
`.env`'s `DSN` still points at `127.0.0.1:3306` for this to work.

### 4. Build the image

```bash
make docker-build
```

Builds and tags the image using the current git tag/version, e.g.
`student-rest-api:v0.1.0`.

### 5. Run the API container

```bash
make docker-run
```

This joins the same `student-api-net` network as MySQL and injects config
from `.env.docker` at runtime - nothing is baked into the image itself.

Verify it's working:

```bash
curl http://localhost:8888/healthcheck
```

### Stopping everything

```bash
make docker-mysql-down
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/v1/students | Get all students |
| GET | /api/v1/student/:id | Get a student by ID |
| POST | /api/v1/students | Create a new student |
| PUT | /api/v1/student/:id | Update a student by ID |
| DELETE | /api/v1/student/:id | Delete a student by ID |
| GET | /healthcheck | Check API health |

## Sample Request Body

Use this JSON body for creating or updating a student:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 20,
  "class": "10th",
  "department": "CS"
}
```

## Example Requests

### Get all students

```bash
curl http://localhost:8888/api/v1/students
```

### Create a student

```bash
curl -X POST http://localhost:8888/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 20,
    "class": "10th",
    "department": "CS"
  }'
```
