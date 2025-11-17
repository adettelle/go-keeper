DB="postgresql://postgres:password@localhost:5433/praktikum-fin?sslmode=disable"
include .env
export 

migrate-create:
	./migrate create -ext sql -dir internal/migrator/migration/ -seq $(name)

migrate-up:
	./migrate -path internal/migrator/migration/ -database $(DB) -verbose up

migrate-down:
	./migrate -path internal/migrator/migration/ -database $(DB) -verbose down

test: prepare-env
	go test ./...

lint: 
	golangci-lint run

vet:
	staticcheck ./...

check: lint vet test

testcov: test
	go test -v -coverpkg=./... -coverprofile=profile.cov ./... && go tool cover -func profile.cov && go tool cover -html profile.cov

testcov2: test
	go test -v -coverpkg=./... -coverprofile=profile.cov ./... && \
	cat profile.cov | grep -v "mock_.*.go" > cover.out && \
	go tool cover -func cover.out && go tool cover -html cover.out

prepare-env:
	docker-compose up minio_mc && docker-compose up -d postgres minio

run-server: prepare-env
	go run ./cmd/server/

run-all:
	docker-compose up --build

generate-client-certs:
	mkdir ./keys1 && \
		openssl req -x509 -newkey rsa:4096 -keyout ./keys1/client_privatekey.pem \
			-out ./keys1/client_cert.pem -sha256 -days 365 -nodes \
			-subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"
