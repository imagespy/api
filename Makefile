DATABASE_ADDR ?= 127.0.0.1:33306
DATABASE_PASS ?= root
DATABASE_USER ?= root

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
	./migrate -source file://./store/gorm/migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASS}@tcp(${DATABASE_ADDR})/imagespy?charset=utf8&parseTime=True&loc=UTC" up

download_migrate:
	curl -L -o migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.0.2/migrate.darwin-amd64.tar.gz
	tar -xvzf migrate.tar.gz
	mv migrate.darwin-amd64 migrate
	rm migrate.tar.gz

test:
	go test -v ./...

test_scrape:
	go test -v ./scrape
