.DEFAULT_GOAL=test

dev_database:
	cd ./dev && docker-compose up -d database

dev_database_rm:
	cd ./dev && docker-compose stop database && docker-compose rm -f database

dev_registry:
	cd ./dev && docker-compose up -d registry

dev_registry_rm:
	cd ./dev && docker-compose stop registry && docker-compose rm -f registry

download_migrate:
	curl -L -o migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.0.2/migrate.darwin-amd64.tar.gz
	tar -xvzf migrate.tar.gz
	mv migrate.darwin-amd64 migrate
	rm migrate.tar.gz

test:
	go test ./...
