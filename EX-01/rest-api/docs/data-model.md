# Data Model

← [Back to README](../README.md)

This page documents the `Student` data model. It covers each field, its type, and its constraints.

The `Student` struct ([models/studentModel.go](../models/studentModel.go)) is both the GORM model and the source of the `students` table schema:

| Field | Type | DB constraints | Client-settable? |
|---|---|---|---|
| `id` | uint | primary key, auto-increment | No: assigned by the database |
| `created_at` | time.Time | — | No: set automatically by GORM on insert |
| `updated_at` | time.Time | — | No: set automatically by GORM on every update |
| `deleted_at` | time.Time (nullable) | indexed | No: see note below |
| `name` | string | not null, max 255 chars | Yes: required |
| `email` | string | unique, not null, max 255 chars | Yes: required, must be a valid email |
| `age` | int | not null | Yes: required, must be > 0 |
| `class` | string | not null, max 255 chars | Yes: required |
| `department` | string | not null, max 255 chars | Yes: required |

Note: Deleting a student removes the row completely. GORM's `gorm.DeletedAt` type can mark a row as deleted instead of removing it. The row then stays in the table, hidden from normal queries. This app does not use that option. It instead uses delete logic that removes the row from the table directly.

Because of this, `deleted_at` is always `null` in this API's responses. A deleted student is gone, not hidden.

