package gen 


//go:generate go run ./cmd/tools/tern.go
//go:generate sqlc generate -f ./internal/store/pgstore/sqlc.yml