.PHONY: test
test:
	go test -count=1 -v ./...

.PHONY: vendor
vendor:
	go mod tidy && go mod vendor

.PHONY: test-e2e-down
test-e2e-down:
	docker-compose -f docker-compose-e2e.yml down

.PHONY: test-e2e
test-e2e: test-e2e-down
	docker-compose -f docker-compose-e2e.yml up -d
	( \
		ING_E2E_KAFKA_BROKER_ADDR=localhost:9092 \
		ING_E2E_POSTGRES_ADDR=localhost:5432 \
		ING_E2E_POSTGRES_USER=ing_user \
		ING_E2E_POSTGRES_PASS=ing_pass \
		ING_E2E_POSTGRES_DB=ing \
		ING_E2E_REDIS_ADDRS=localhost:6379,localhost:6380,localhost:6381 \
		ING_E2E_REDIS_USER=ing_user \
		ING_E2E_REDIS_PASS=ing_pass \
			go test -count=1 -v -tags=e2e ./... \
		|| \
		docker-compose -f docker-compose-e2e.yml down \
	)
	docker-compose -f docker-compose-e2e.yml down
