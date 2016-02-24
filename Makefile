test: vet lint
	go test

vet:
	go vet ./...

lint:
	! golint ./... | grep -v 'should have comment'

.PHONY: test vet lint
