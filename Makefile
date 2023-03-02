test:
	go test -v ./...

test-e2e-down:
	docker-compose -f docker-compose-e2e.yml down

test-e2e: test-e2e-down
	docker-compose -f docker-compose-e2e.yml up -d
	( ING_E2E_KAFKA_BROKER_ADDR=localhost:9092 go test -v -tags=e2e ./... || docker-compose -f docker-compose-e2e.yml down )
	docker-compose -f docker-compose-e2e.yml down
