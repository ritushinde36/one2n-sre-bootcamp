# API Reference

← [Back to README](../README.md)

This page documents every endpoint the API exposes. It covers request and response formats, pagination, and status codes. Use it as a reference when you call the API.

Base path for student resources: `/api/v1`

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/students` | List students (paginated) |
| GET | `/api/v1/student/:id` | Get a single student by ID |
| POST | `/api/v1/students` | Create a new student |
| PUT | `/api/v1/student/:id` | Update a student by ID |
| DELETE | `/api/v1/student/:id` | Delete a student by ID |
| GET | `/healthcheck` | Liveness probe (confirms the process is running) |
| GET | `/readyz` | Readiness probe (confirms the app can reach its database) |

## Pagination

`GET /api/v1/students` accepts two optional query parameters:

| Param | Meaning | Default |
|---|---|---|
| `limit` | Maximum number of students to return in one response | 50 |
| `offset` | Number of students to skip from the start of the list, before `limit` applies | 0 |

For example, `limit=10&offset=20` skips the first 20 students and returns the next 10 (students 21–30). This fetches "page 3" when each page has 10 students:

```bash
curl "http://localhost:8888/api/v1/students?limit=10&offset=20"
```

## Request body (Create / Update)

Both `POST /api/v1/students` and `PUT /api/v1/student/:id` take the same JSON shape. All fields are required. **The API rejects unknown fields with a 400 response.**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 20,
  "class": "10th",
  "department": "CS"
}
```

| Field | Type | Validation |
|---|---|---|
| `name` | string | required |
| `email` | string | required, must be a valid email, must be unique across students |
| `age` | int | required, must be > 0 |
| `class` | string | required |
| `department` | string | required |

The API always returns errors as `{"error": "<message>"}`.

## Status codes

| Status | When |
|---|---|
| 200 OK | Successful GET / PUT / DELETE |
| 201 Created | Successful POST |
| 400 Bad Request | Malformed JSON, missing or invalid required fields, unknown fields in the body, non-numeric `:id` |
| 404 Not Found | No student exists with the given `:id` |
| 409 Conflict | `email` collides with an existing student |
| 500 Internal Server Error | Unexpected database failure |
| 503 Service Unavailable | `/readyz` only. Database is unreachable |

## Example requests

The student ID `1` below is only an example. Replace it with the actual numeric ID of the student to look up, update, or delete.

**Get all students**

```bash
curl http://localhost:8888/api/v1/students
```

**Get one student**

```bash
curl http://localhost:8888/api/v1/student/1
```

**Create a student**

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

**Update a student**

```bash
curl -X PUT http://localhost:8888/api/v1/student/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 21,
    "class": "11th",
    "department": "CS"
  }'
```

**Delete a student**

```bash
curl -X DELETE http://localhost:8888/api/v1/student/1
```

**Health / readiness**

```bash
curl http://localhost:8888/healthcheck
curl http://localhost:8888/readyz
```
