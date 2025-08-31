# squads-rest-api

🚀 一个基于 Golang + SQLite3 的 RESTful API 服务，支持 Multisig / Vault / Member 三个资源的完整 CRUD 操作，支持分页、搜索、过滤、排序，并提供子资源查询。


## 目录

- 数据库表结构 (SQLite3 Schema)
- 功能特性
- 快速开始
- 健康检查
- 通用查询参数说明
- API 文档
  - Multisigs
  - Vaults
  - Members
  - 子资源查询
- 示例响应 JSON
- 错误响应格式
- License

## 数据库表结构 (SQLite3 Schema)

```
-- Multisig 表
CREATE TABLE multisig (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    multisig_address TEXT NOT NULL UNIQUE,
    name TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Vault 表
CREATE TABLE vault (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_address TEXT NOT NULL UNIQUE,
    multisig_address TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (multisig_address) REFERENCES multisig(multisig_address)
);

-- Member 表
CREATE TABLE member (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    member_address TEXT NOT NULL UNIQUE,
    name TEXT,
    multisig_address TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (multisig_address) REFERENCES multisig(multisig_address)
);
```

## 功能特性

- Multisig / Vault / Member 三个表的 CRUD API
- 通用查询参数：分页、搜索、排序
- 子资源接口：/multisigs/{id}/vaults、/multisigs/{id}/members
- 标准 JSON 响应
- 健康检查接口 /health


## 快速开始

```
# 克隆项目
git clone https://github.com/yourname/squads-rest-api.git
cd squads-rest-api

# 构建
make build

# 运行
./squads-rest-api

# 服务默认运行在：
http://localhost:8080
```

## 健康检查

curl -s http://localhost:8080/health | jq

响应：
{
  "status": "ok"
}

## 通用查询参数说明

参数名       | 类型   | 说明                                           | 示例
------------|--------|-----------------------------------------------|------
page        | int    | 页码，从 1 开始（默认：1）                    | ?page=2
page_size   | int    | 每页数量（默认：10）                          | ?page_size=20
search      | string | 模糊搜索，作用于部分字段（如 name、description、address 等） | ?search=alice
sort        | string | 排序字段和顺序，格式：字段:asc 或 字段:desc   | ?sort=created_at:desc
组合查询     | -      | 参数可组合使用                                | ?page=1&page_size=5&search=test&sort=updated_at:asc


## API 文档

```
Multisigs

创建
curl -s -X POST http://localhost:8080/multisigs \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc123","name":"Squad A","description":"First squad"}' | jq

查询列表
curl -s "http://localhost:8080/multisigs?page=1&page_size=10&sort=created_at:desc" | jq

查询单个
curl -s http://localhost:8080/multisigs/1 | jq

更新
curl -s -X PUT http://localhost:8080/multisigs/1 \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc999","name":"Squad A Updated","description":"Updated desc"}' | jq

删除
curl -s -X DELETE http://localhost:8080/multisigs/1 | jq

Vaults

创建
curl -s -X POST http://localhost:8080/vaults \
  -H "Content-Type: application/json" \
  -d '{"vault_address":"0xvault1","multisig_address":"0xabc123"}' | jq

查询列表
curl -s "http://localhost:8080/vaults?page=1&page_size=10" | jq

查询单个
curl -s http://localhost:8080/vaults/1 | jq

更新
curl -s -X PUT http://localhost:8080/vaults/1 \
  -H "Content-Type: application/json" \
  -d '{"vault_address":"0xvault99","multisig_address":"0xabc123"}' | jq

删除
curl -s -X DELETE http://localhost:8080/vaults/1 | jq

Members

创建
curl -s -X POST http://localhost:8080/members \
  -H "Content-Type: application/json" \
  -d '{"member_address":"0xmem1","name":"Alice","multisig_address":"0xabc123"}' | jq

查询列表
curl -s "http://localhost:8080/members?page=1&page_size=10" | jq

查询单个
curl -s http://localhost:8080/members/1 | jq

更新
curl -s -X PUT http://localhost:8080/members/1 \
  -H "Content-Type: application/json" \
  -d '{"member_address":"0xmemX","name":"Alice Updated","multisig_address":"0xabc123"}' | jq

删除
curl -s -X DELETE http://localhost:8080/members/1 | jq

子资源查询

获取某个 Multisig 下的 Vaults
curl -s http://localhost:8080/multisigs/1/vaults | jq

获取某个 Multisig 下的 Members
curl -s http://localhost:8080/multisigs/1/members | jq
```

## 示例响应 JSON

```
Multisigs 列表（带分页信息）
{
  "page": 1,
  "page_size": 10,
  "total": 25,
  "items": [
    {
      "id": 1,
      "multisig_address": "0xabc123",
      "name": "Squad A",
      "description": "First squad",
      "created_at": "2025-08-31T08:00:00Z",
      "updated_at": "2025-08-31T08:00:00Z"
    },
    {
      "id": 2,
      "multisig_address": "0xdef456",
      "name": "Squad B",
      "description": "Second squad",
      "created_at": "2025-08-31T09:00:00Z",
      "updated_at": "2025-08-31T09:00:00Z"
    }
  ]
}

Vaults 列表（带分页信息）
{
  "page": 1,
  "page_size": 10,
  "total": 5,
  "items": [
    {
      "id": 1,
      "vault_address": "0xvault1",
      "multisig_address": "0xabc123",
      "created_at": "2025-08-31T08:30:00Z",
      "updated_at": "2025-08-31T08:30:00Z"
    }
  ]
}

Members 列表（带分页信息）
{
  "page": 1,
  "page_size": 10,
  "total": 8,
  "items": [
    {
      "id": 1,
      "member_address": "0xmem1",
      "name": "Alice",
      "multisig_address": "0xabc123",
      "created_at": "2025-08-31T08:15:00Z",
      "updated_at": "2025-08-31T08:15:00Z"
    },
    {
      "id": 2,
      "member_address": "0xmem2",
      "name": "Bob",
      "multisig_address": "0xdef456",
      "created_at": "2025-08-31T08:45:00Z",
      "updated_at": "2025-08-31T08:45:00Z"
    }
  ]
}

错误响应格式
{
  "error": "not found"
}
```

## License

MIT License.
