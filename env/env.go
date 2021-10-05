package env

// 生产环境
// 通过 container_name:port 方式访问其他容器
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
