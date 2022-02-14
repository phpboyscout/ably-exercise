protoc:
	protoc -I proto --go_out=go/pkg --go-grpc_out=go/pkg --go_opt=paths=import --go-grpc_opt=paths=import ./proto/server.proto

build-server:
	cd go && go mod download
	go build -ldflags="-w -s -extldflags '-static'" -o bin/server -a go/cmd/server/main.go

build-client-go:
	cd go && go mod download
	cd go && go build -ldflags="-w -s -extldflags '-static'" -o bin/client -a go/cmd/client/main.go

test:
	cd go && staticcheck ./...
	cd go && go test -cover  -coverprofile=coverage.out ./...

coverage:
	cd go && go tool cover -html=coverage.out

run:
	cd go && go run cmd/server/main.go 9090
