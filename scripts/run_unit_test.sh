cd ../service
go test -tags="unit" -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out