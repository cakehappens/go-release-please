.PHONY: gen
gen:
	cd internal/github && go run github.com/Khan/genqlient

.PHONY: build
build:
	go build -o go-release-please ./cmd/go-release-please/main.go