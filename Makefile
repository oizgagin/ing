test:
	go test -count=1 -v ./...

test-e2e-down:
	docker-compose -f docker-compose-e2e.yml down

test-e2e: test-e2e-down
	docker-compose -f docker-compose-e2e.yml up -d
	( \
		ING_E2E_KAFKA_BROKER_ADDR=localhost:9092 \
		ING_E2E_POSTGRES_USER=ing_user \
		ING_E2E_POSTGRES_PASS=ing_pass \
		ING_E2E_POSTGRES_DB=ing \
			go test -count=1 -v -tags=e2e ./... \
		|| \
		docker-compose -f docker-compose-e2e.yml down \
	)
	docker-compose -f docker-compose-e2e.yml down
