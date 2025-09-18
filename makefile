rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

server-log:
	docker-compose logs server --follow  

db-log:
	docker-compose logs postgres_new --follow