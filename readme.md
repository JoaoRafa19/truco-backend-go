
![header](https://capsule-render.vercel.app/api?type=venom&color=auto&height=400&section=header&text=Truco&fontSize=90&rotate=10)

![go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)![pgsql](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)![flutter](https://img.shields.io/badge/Flutter-235997?style=for-the-badge&logo=flutter&logoColor=white)

# Truco

API de truco online para applicação mobile





Utiliza `sqlc` para gerar as interfaces das entidades das tabelas dos bancos de dados (não é um ORM) e as queries SQL.
Utiliza o `tern` para criar e executar as migations.

## Go generate

Executa os comandos declarados em `gen.go`
```go
package gen 


//go:generate go run ./cmd/tools/terndotenv/main.go
//go:generate sqlc generate -f ./internal/store/pgstore/sqlc.yml
```
```shell
go generate ./...
```

## Migrations
Utiliazando o tern para criar migrações, mas para executar com o ambiente local do docker pelo arquivo .env
utiliza o `os\exec` do go para rodar comandos no ambiente

```shell
go run ./cmd/tools/terndotenv/main.go
```

## Queries

Usa `sqlc` para gerar as queries

```shell
sqlc generate -f ./internal/store/pgstore/sqlc.yml
```


## Deps


#### Install all deps:
```shell
go mod tidy
```

- **tern**
```shell
 go install github.com/jackc/tern/v2@latest
 ```

- **sqlc**
```shell
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```



## Generate

# Requisitos

## Frontend

- [ ]  Tela de login
- [ ]  Tela de entrar / criar sala
- [ ] Tela de jogo (bonfire)


### Backend

- [ ] Salva o estado atual da sala e do deck
- [ ] Cada sala tem o estado do jogo
- [ ] Regras do jogo
- [ ] Cada sala tem um deck
- [ ] Regras de pontuaçao 
- [X] Setup banco de dados para salas de jogo
- [X] Setup banco de dados para jogadores
- [ ] Autenticação JWT?
- [X] Criar sala de jogo
- [X] Entrar na sala
- [X] Receber mensagens do Websocket
- [X] Sai da sala e remove a sala caso seja a ultima conexão

Cada sala tem os jogadores salvos