.PHONY: default install-opencv 

.DEFAULT_GOAL := default


default: build

build: 
	go build ./cmd/daemon

dev-api:
	go run ./cmd/api


# Build the Media MTX Docker image for the test environment
# and run it locally	
dev-env:
	docker build -t sscs-mtx ./dev

	docker run -d --rm --network=host -p 8554:8554 -p 8889:8889 \
	-v ./dev/mediamtx.yml --name sscs-mtx sscs-mtx

	docker run -d --rm --network=host  -p 5432:5432   \
	-e POSTGRES_PASSWORD=gorm -e POSTGRES_USER=gorm  \
	-e  POSTGRES_DB=gorm --name sscs-postgres postgres

# it uses the openCV Makefile to install openCV in the system easily
install-opencv:
	make -f Makefile.opencv install


# runs a keycloak instankce to simulate jwt authentication. ATTENTION! this is only to be used in
# a development environment, NOT in production
key-cloak: 
	docker run --name=keycloak -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:23.0.7 start-dev

