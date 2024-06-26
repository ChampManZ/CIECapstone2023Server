include .env
export

.PHONY: sql setup run stop del run-in-background terminate

sql:
	docker compose -f ${MYSQL_COMPOSE} up -d
	@echo "Waiting for MySQL to be ready..."
	@until docker exec $(MYSQL_CONTAINER_NAME) mysql -h localhost -u$(MYSQL_USER) --password=$(MYSQL_PASSWORD) -e "SELECT 1" &> /dev/null; do \
	    sleep 1; \
	done
	@echo "MySQL is up"

setup:
	docker exec -i $(MYSQL_CONTAINER_NAME) mysql -u$(MYSQL_USER) --password=$(MYSQL_PASSWORD) $(MYSQL_DBNAME) < $(SQL_FILE)
	go mod tidy

run:
	docker compose -f ${MYSQL_COMPOSE} up -d
	go run main.go
stop:
	docker compose -f ${MYSQL_COMPOSE} down
del:
	docker compose -f ${MYSQL_COMPOSE} down -v
build:
	GOOS=linux GOARCH=amd64 go build -o capstone

run-in-background:
	@echo "Starting capstone in background..."
	@nohup ./capstone &

terminate:
	@echo "Terminating capstone..."
	@PIDS=$$(pgrep -f './capstone') ; \
	if [ -n "$$PIDS" ]; then \
		kill $$PIDS ; \
	else \
		echo "No capstone process found."; \
	fi
