
all: run

run:
	go install && aviasales_test
test:
	go test -v ./...
up_env:
	docker-compose up
#test request
load_test:
	curl -v localhost:9090/load -d '["foobar", "BarFoo", "aabb", "baba", "test"]'
get_test:
	curl -v localhost:9090/get?word=foobar