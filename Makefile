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

test_scrape:
	RUN_SCRAPE_TESTS=1 go test -v ./scrape

build:
	go build -ldflags="-X github.com/imagespy/api/version.Version=${VERSION}"

build_docker:
	docker build --build-arg VERSION=${VERSION} -t imagespy/api:${VERSION} .

release_docker: build_docker
	docker push imagespy/api:${VERSION}
