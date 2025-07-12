setup-local-tests:
	docker compose -f 'docker-compose-tests.yml' up -d

teardown-local-tests:
	docker compose -f 'docker-compose-tests.yml' down

tests:
	go test -v ./...
