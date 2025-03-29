cert:
	mkdir -p config/certs
	openssl req -x509 -nodes -newkey rsa:2048 \
	  -keyout config/certs/server.key -out config/certs/server.crt \
	  -days 365 \
	  -subj "/CN=opamp-backend" \
	  -addext "subjectAltName = DNS:opamp-backend, DNS:host.docker.internal, DNS:localhost"

build:
	go build -o opamp-server ./cmd/server.go

docker:
	docker build -t opamp-backend .

run:
	./opamp-server