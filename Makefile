postgres:
# Start a PostgreSQL container with port mapping (5432:5432) 
# and set `root` as the superuser. PostgreSQL automatically creates a `root` database for the superuser.
# 使用 POSTGRES_USER=root 启动容器，PostgreSQL 会自动创建一个与「超级用户」同名的数据库 root
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
# Create a database `simple_bank` with `root` as the owner.
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
	
migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
# '1' means only applying the next migration version from the current one

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
	
migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
# '1' means only applying the last down migration version that was applied before

# migrateup 和 migratedown 通过 localhost:5432 访问 PostgreSQL
# 如果没有端口映射，主机无法连接到容器内的 PostgreSQL 服务，因此迁移命令会失败

sqlc:
	sqlc generate

test:
	go test -v -cover ./...
	
server:
	go run main.go
	
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/Oliver-Zen/simplebank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock