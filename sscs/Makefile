.PHONY: default install-opencv 

.DEFAULT_GOAL := default


default: build

build: 
	go build


# Build the Media MTX Docker image for the test environment
# and run it locally	
dev-env:
	docker build -t sscs-mtx ./dev

	docker run -d --rm --network=host -p 8554:8554 -p 8889:8889 \
	-v ./dev/mediamtx.yml --name sscs-mtx sscs-mtx

	docker run -d --rm --network=host --name sscs-postgres \
	-e POSTGRES_PASSWORD=gorm -e POSTGRES_USER=gorm -e POSTGRES_DB=gorm -p 5432:5432 postgres

# it uses the openCV Makefile to install openCV in the system easily
install-opencv:
	make -f Makefile.opencv install


