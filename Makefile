test: imports vet
	go test

imports:
	goimports -w .

vet:
	go vet ./...

lint:
	! golint ./... | grep -v 'should have comment'

.PHONY: test imports vet lint
