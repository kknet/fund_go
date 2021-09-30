package env

// 部署环境
// 1. 通过 container_name:port 方式访问其他容器
// 2. 容器内访问宿主机 host.docker.internal
const (
	RedisHost    = "fund_redis:6379"
	MongoHost    = "fund_mongo:27017"
	PostgresHost = "fund_postgres:5432"
)

// 本地环境
const (
//RedisHost    = "localhost:6379"
//MongoHost    = "localhost:27017"
//PostgresHost = "localhost:5432"
)
