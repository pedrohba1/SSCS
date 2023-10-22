.PHONY: default install-opencv 

.DEFAULT_GOAL := default


default: build

build: 
	go build


# Build the Docker image for the test environment
# and run it locally	
dev-env:
	docker build -t sscs-dev-env ./dev
	docker run --name sscs-dev-env-container -d sscs-dev-env






# it uses the openCV Makefile to install openCV in the system easily
install-opencv:
	make -f Makefile.opencv install


