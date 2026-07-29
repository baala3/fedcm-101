.PHONY: run-idp run-sp tidy clean

run-idp:
	go run ./cmd/idp

run-sp:
	go run ./cmd/sp

tidy:
	go mod tidy

clean:
	rm -f data/idp.db
