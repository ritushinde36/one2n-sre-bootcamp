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

## Setup

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
