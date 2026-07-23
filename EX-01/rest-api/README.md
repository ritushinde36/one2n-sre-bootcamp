# Student REST API

A simple Go-based REST API for doing basic CRUD operations on student records using Gin Gonic, GORM, and MySQL.

## Features

- Create, read, update, and delete student records
- MySQL database integration with go using GORM
- Router set up using Gin and Gonic.
- Automatic table creation for the Student model struct
- API served on port 8888

## Project Structure

- main.go - application entry point. Connects to DB, creates the table and route setup. 
- controllers/StudentController.go - request handlers for CRUD operations
- models/studentModel.go - Student data model
- connections/db_connection.go - MySQL connection and database migration

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

3. Run the server:

```bash
go run main.go
```

The API will start on:

```text
http://localhost:8888
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /students | Get all students |
| GET | /student/:id | Get a student by ID |
| POST | /students | Create a new student |
| PUT | /student/:id | Update a student by ID |
| DELETE | /student/:id | Delete a student by ID |

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
curl http://localhost:8888/students
```

### Create a student

```bash
curl -X POST http://localhost:8888/students \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 20,
    "class": "10th",
    "department": "CS"
  }'
```
