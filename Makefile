DATABASE_ADDR ?= 127.0.0.1:33306
DATABASE_CREDENTIALS ?= root:root
MIGRATE_OS_ARCH ?= darwin-amd64
MIGRATE_VERSION = v4.0.2
VERSION ?= master

.DEFAULT_GOAL=test

dev_database:
	cd ./dev && docker-compose up -d database

dev_database_rm:
	cd ./dev && docker-compose stop database && docker-compose rm -f database

dev_registry:
	cd ./dev && docker-compose up -d registry

dev_registry_rm:
	cd ./dev && docker-compose stop registry && docker-compose rm -f registry

dev_migrate:
	./migrate -source file://./store/gorm/migrations -database "mysql://${DATABASE_CREDENTIALS}@tcp(${DATABASE_ADDR})/imagespy?charset=utf8&parseTime=True&loc=UTC" up

download_migrate:
	curl -L -o migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${MIGRATE_OS_ARCH}.tar.gz
	tar -xvzf migrate.tar.gz
	mv migrate.${MIGRATE_OS_ARCH} migrate
	rm migrate.tar.gz

generate_mocks:
	mockgen -source=./scrape/scraper.go -destination=./scrape/mock.go -package scrape
	mockgen -source=./store/store.go -destination=./store/mock/mock.go -package mock

test:
	go test -v ./...

test_e2e:
	go get github.com/DATA-DOG/godog/cmd/godog
	cd ./e2e && godog

build:
	go build -ldflags="-X github.com/imagespy/api/version.Version=${VERSION}"

build_docker:
	docker build --build-arg VERSION=${VERSION} -t imagespy/api:${VERSION} .

release_docker: build_docker
	docker push imagespy/api:${VERSION}

setup_e2e:
	docker pull golang@sha256:b35aec8702783621fbc0cd08cbc6a8fa8ade8b7233890f3a059645f3b2cfa93f
	docker tag golang@sha256:b35aec8702783621fbc0cd08cbc6a8fa8ade8b7233890f3a059645f3b2cfa93f 127.0.0.1:52854/golang:1.12.3
	docker pull golang@sha256:83e8267be041b3ddf6a5792c7e464528408f75c446745642db08cfe4e8d58d18
	docker tag golang@sha256:83e8267be041b3ddf6a5792c7e464528408f75c446745642db08cfe4e8d58d18 127.0.0.1:52854/golang:1.12.4
	docker tag golang@sha256:83e8267be041b3ddf6a5792c7e464528408f75c446745642db08cfe4e8d58d18 127.0.0.1:52854/golang:latest
	docker pull debian@sha256:65e581e00438a33ccbb0bd2d74f03de99d8bec8abca982e906a055f828bc5b57
	docker tag debian@sha256:65e581e00438a33ccbb0bd2d74f03de99d8bec8abca982e906a055f828bc5b57 127.0.0.1:52854/debian:stretch-20190326
