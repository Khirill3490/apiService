export


run-service:
	@echo "Starting API Service..."
	@CONFIG_PATH=$$CONFIG_PATH go run main.go
	
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-version:
	migrate -path migrations -database "$$DATABASE_URL" version

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
