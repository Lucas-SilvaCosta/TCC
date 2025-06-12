run-vendor:
	go mod vendor

undo-vendor:
	rm -rf vendor

build-local: run-vendor build undo-vendor

test-local: run-vendor test undo-vendor

build:
	go build -v .

test:
	go test -v .

clean-vendor:
	rm -rf vendor
clean: clean-vendor
	rm -rf simplestate
