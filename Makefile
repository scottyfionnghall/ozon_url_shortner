build:
	@go build -o ./bin/ozonshrt ./cmd/api  

run: build
	@./bin/ozonshrt

# test:
# 	@go test -v ./...