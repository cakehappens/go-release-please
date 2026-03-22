.PHONY: gen
gen:
	cd internal/github && go run github.com/Khan/genqlient
