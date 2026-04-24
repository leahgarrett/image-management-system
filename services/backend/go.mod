module github.com/leahgarrett/image-management-system/services/backend

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.5.5
	golang.org/x/crypto v0.21.0
)

require github.com/sqlc-dev/pqtype v0.3.0 // indirect
