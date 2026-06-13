# zredis

一个用 Go 从零构建的 Redis 克隆，用于学习 Redis 内部原理。

A Redis clone built from scratch in Go, for learning Redis internals.

## 快速开始

```bash
# 启动服务器
go run ./cmd/server/

# 另开终端，用 redis-cli 连接
redis-cli
```

## 已实现

| 分类   | 命令                                                         |
|--------|--------------------------------------------------------------|
| 通用   | PING, DEL, EXISTS, KEYS, FLUSHALL, EXPIRE, TTL, PERSIST     |
| String | SET, GET, INCR, INCRBY, DECR, DECRBY, APPEND, STRLEN, MGET, MSET, GETSET, SETEX, SETNX |
| Hash   | HSET, HGET, HDEL, HGETALL, HEXISTS, HLEN, HKEYS, HVALS      |
| List   | LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, LINDEX, LSET        |
| Set    | SADD, SREM, SMEMBERS, SISMEMBER, SCARD, SINTER, SUNION, SDIFF |
| ZSet   | ZADD, ZRANGE, ZRANK, ZSCORE, ZREM, ZCARD                    |

**共 40 个命令，172 个测试用例，race detector 通过。**

## 架构

```
cmd/server/main.go     ← 启动入口
core/
├── value.go            RESP 5 种值类型（Simple String / Error / Integer / Bulk String / Array）
├── store.go            通用引擎（DataEntry、过期、删除）
├── ttl.go              TTL 管理（Expire、TTL、Persist、CleanExpired）
├── string_cmd.go       String 命令
├── hash.go             Hash 命令
├── list.go             List 命令（基于 container/list）
├── set.go              Set 命令
├── skiplist.go         跳表实现（Redis 风格，含 span 支持 O(log n) 排名查询）
├── zset.go             ZSet 命令
├── resp.go             RESP 协议解码器 + 编码器
├── dispatch.go         40 个命令分发器
├── server.go           TCP 服务器
```

## 学习文档

按照构建顺序，共 12 步，每步一篇中文讲解：

| 步骤 | 文档 | 内容 |
|------|------|------|
| 1 | [docs/01-value-types.md](docs/01-value-types.md) | RESP 5 种值类型 |
| 2 | [docs/02-thread-safe-store.md](docs/02-thread-safe-store.md) | 线程安全 Store |
| 3 | [docs/03-ttl-expiration.md](docs/03-ttl-expiration.md) | TTL 与过期机制 |
| 4 | [docs/04-string-commands.md](docs/04-string-commands.md) | String 命令 |
| 5 | [docs/05-hash.md](docs/05-hash.md) | Hash 命令 |
| 6 | [docs/06-list.md](docs/06-list.md) | List 命令 |
| 7 | [docs/07-set.md](docs/07-set.md) | Set 命令 |
| 8 | [docs/08-zset.md](docs/08-zset.md) | ZSet + 跳表 |
| 9 | [docs/09-resp-decoder.md](docs/09-resp-decoder.md) | RESP 解码器 |
| 10 | [docs/10-resp-encoder.md](docs/10-resp-encoder.md) | RESP 编码器 |
| 11 | [docs/11-dispatcher.md](docs/11-dispatcher.md) | 命令分发器 |
| 12 | [docs/12-tcp-server.md](docs/12-tcp-server.md) | TCP 服务器 |

## 协议

完整实现 [RESP 协议](https://redis.io/docs/latest/develop/reference/protocol-spec/)，可与标准 `redis-cli` 互操作。

## 运行测试

```bash
go test ./core/ -race -v
```

## License

MIT
