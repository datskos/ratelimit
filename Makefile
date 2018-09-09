run: build
	./bin/ratelimit

build:
	go build -o bin/ratelimit ./cmd/ratelimit

test:
	go test ./pkg...

proto:
	protoc --go_out=plugins=grpc:. pkg/proto/*.proto
