# Tech Stack

← [Back to README](../README.md)

This page lists the libraries and tools this project uses. It groups them by what each one is for.

| Concern | Library |
|---|---|
| HTTP router | [gin-gonic/gin](https://github.com/gin-gonic/gin) |
| ORM | [gorm.io/gorm](https://gorm.io/) + [gorm.io/driver/mysql](https://github.com/go-gorm/mysql) |
| DB driver | [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) |
| Migrations | [pressly/goose](https://github.com/pressly/goose) |
| Config | [joho/godotenv](https://github.com/joho/godotenv) (loads `.env`) |
| Logging | Go standard library `log/slog` (JSON handler) |
| Test assertions | [stretchr/testify](https://github.com/stretchr/testify) |
| Test infrastructure | [testcontainers-go](https://github.com/testcontainers/testcontainers-go) |
| Containerization | [Docker](https://www.docker.com/) (multi-stage [Dockerfile](../Dockerfile)) |
| Container orchestration (local) | [Docker Compose](https://docs.docker.com/compose/) ([docker-compose.yml](../docker-compose.yml)) |

Go version: see [go.mod](../go.mod) (currently 1.26.5).
