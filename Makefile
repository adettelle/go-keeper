DB="postgresql://postgres:password@localhost:5433/praktikum-fin?sslmode=disable"

migrate-create:
	./migrate create -ext sql -dir internal/migrator/migration/ -seq $(name)

migrate-up:
	./migrate -path internal/migrator/migration/ -database $(DB) -verbose up

migrate-down:
	./migrate -path internal/migrator/migration/ -database $(DB) -verbose down