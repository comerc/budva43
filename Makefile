submodule:
	git submodule update --init

lint:
	golangci-lint run

test:
	go test -v ./...

build:
	go build -o bin/app main.go

run:
	go run main.go

clean:
	rm -f bin/app

all-build:
	lint
	test
	clean
	build

