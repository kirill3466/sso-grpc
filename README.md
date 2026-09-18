### Миграции

```powershell
cd D:\ne_udalyat\go\grpc\authyt\sso
go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations
```

### SSO

```powershell
cd D:\ne_udalyat\go\grpc\authyt\sso
go run ./cmd/sso --config=./config/local.yaml
```

### RTDB

```powershell
cd D:\ne_udalyat\go\grpc\authyt\rtdb
go run ./cmd/rtdb --config=./config/local.yaml
```

### прото генерация

```powershell
cd D:\ne_udalyat\go\grpc\authyt\protos
make gen_grpc_sso
make gen_grpc_rtdb
```
