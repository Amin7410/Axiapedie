.PHONY: build run dev clean db-up

APP_NAME=axia-wiki
DB_PATH=./data/wiki.db

build:
	go build -tags "fts5" -o bin/${APP_NAME} ./cmd/server/main.go

run: build
	./bin/${APP_NAME}

dev:
	go run -tags "fts5" ./cmd/server/main.go

clean:
	rm -rf bin/

# Giả lập lệnh tạo database
db-up:
	mkdir -p data
	sqlite3 ${DB_PATH} < ../Tài\ liệu\ chính\ thức/database_schema.sql
