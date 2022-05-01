.PHONY: init
init: tools test

.PHONY: tools
tools:
	go install github.com/ArcanjoQueiroz/wait-for@v0.0.3

.PHONY: test
test: middleware.up
	echo "TODO"

## mysql ##
mysql.console: middleware.up
	docker compose exec mysql mysql -uroot

## middleware ##
HOST_MYSQL_PORT ?= 3306

.PHONY: middleware.up
middleware.up:
	docker compose --profile=middleware up -d
	wait-for --type=mysql --name=develop --port=$(HOST_MYSQL_PORT) --user=develop --password=develop --seconds=3 --maxAttempts=10

.PHONY: middleware.kill
middleware.kill:
	docker compose --profile=middleware kill
