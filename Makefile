dev_database:
	cd ./dev && docker-compose up -d database

dev_database_rm:
	cd ./dev && docker-compose rm database
