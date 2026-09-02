# Local Setup

← [Back to README](../README.md)

This page explains how to set up and run the app on your machine, without Docker. It covers creating the database, configuring `.env`, and starting the server. To run it in a container instead, see [Docker & Docker Compose](docker.md).

1. **Create the database** in MySQL. Connect to your MySQL server with a username and password (for example, `mysql -u root -p`), then run:

   ```sql
   CREATE DATABASE student_db;
   ```

   Note the username and password you used to connect. You will need them again in step 3, to build the `DSN` value.

2. **Install Go dependencies:**

   ```bash
   make deps
   ```

3. **Configure environment variables.** Copy the example file and fill in your MySQL DSN:

   ```bash
   cp .env.example .env
   ```

   For running the app directly (this section), set `DSN`'s host to `127.0.0.1`, for example:

   ```
   root:yourpassword@tcp(127.0.0.1:3306)/student_db?charset=utf8mb4&parseTime=True&loc=Local
   ```

   You will reuse this same `.env` file if you later switch to [Docker & Docker Compose](docker.md). That flow needs `DSN`'s host set to `student-mysql` instead. See that page for why.

   `.env` loads automatically on startup, and it is gitignored, so your local credentials never get committed. See [Environment Variables](environment-variables.md) for what each key means.

4. **Run database migrations** to create the schema (see [Database Migrations](migrations.md) for details):

   ```bash
   make migrate-up
   ```

5. **Start the server:**

   ```bash
   make run
   ```

   The API listens on this address by default:

   ```
   http://localhost:8888
   ```

   Change the port with `PORT`.

   You can also run the app directly, with `go run .`, once `.env` is in place and migrations are applied.
