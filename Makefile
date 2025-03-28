build:
	go build -o opamp-server ./cmd/server.go

docker:
	docker build -t opamp-backend .

run:
	./opamp-server