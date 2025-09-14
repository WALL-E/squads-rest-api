# Squads Rest API

🚀 一个基于 Golang + MySQL 的 RESTful API 服务，支持 Multisig / Vault / Member / Spend 四个资源的完整 CRUD 操作，支持分页、搜索、过滤、排序。

## 功能特性

- Multisig / Vault / Member / Spend 四个表的 CRUD API
- 通用查询参数：分页、搜索、排序
- 子资源接口：
  - /multisigs/{multisig_address}/vaults
  - /multisigs/{multisig_address}/members
  - /multisigs/{multisig_address}/spends
- 标准 JSON 响应
- 健康检查接口 /health

## 健康检查

```
curl -s http://localhost:8090/health | jq

响应：
{
  "success": true,
  "message": "ok"
}
```


## API 文档

Multisigs
```
创建
curl -s -X POST http://localhost:8090/multisigs \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc123","name":"Squad A","description":"First squad"}' | jq

查询列表
curl -s "http://localhost:8090/multisigs" | jq

查询单个
curl -s http://localhost:8090/multisigs/0xabc123 | jq

更新
curl -s -X PUT http://localhost:8090/multisigs/0xabc123 \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc123","name":"Squad A Updated","description":"Updated desc"}' | jq

删除
curl -s -X DELETE http://localhost:8090/multisigs/0xabc123 | jq
```

Spends
```
创建
curl -s -X POST http://localhost:8090/multisigs/0xabc123/spends \
  -H "Content-Type: application/json" \
  -d '{"spend_address":"0xspend1"}' | jq

查询列表
curl -s "http://localhost:8090/multisigs/0xabc123/spends" | jq

查询单个
curl -s http://localhost:8090/multisigs/0xabc123/spends/0xspend1 | jq

更新
curl -s -X PUT http://localhost:8090/multisigs/0xabc123/spends/0xspend1 \
  -H "Content-Type: application/json" \
  -d '{"spend_address":"0xspend1"}' | jq

删除
curl -s -X DELETE http://localhost:8090/multisigs/0xabc123/spends/0xspend1 | jq
```

Vaults
```
创建
curl -s -X POST http://localhost:8090/multisigs/0xabc123/vaults \
  -H "Content-Type: application/json" \
  -d '{"vault_address":"0xvault1"}' | jq

查询列表
curl -s "http://localhost:8090/multisigs/0xabc123/vaults" | jq

查询单个
curl -s http://localhost:8090/multisigs/0xabc123/vaults/0xvault1 | jq

更新
curl -s -X PUT http://localhost:8090/multisigs/0xabc123/vaults/0xvault1 \
  -H "Content-Type: application/json" \
  -d '{"vault_address":"0xvault1"}' | jq

删除
curl -s -X DELETE http://localhost:8090/multisigs/0xabc123/vaults/0xvault1 | jq
```

Members
```
创建
curl -s -X POST http://localhost:8090/multisigs/0xabc123/members \
  -H "Content-Type: application/json" \
  -d '{"member_address":"0xmember1"}' | jq

查询列表
curl -s "http://localhost:8090/multisigs/0xabc123/members" | jq

查询单个
curl -s http://localhost:8090/multisigs/0xabc123/members/0xmember1 | jq

更新
curl -s -X PUT http://localhost:8090/multisigs/0xabc123/members/0xmember1 \
  -H "Content-Type: application/json" \
  -d '{"member_address":"0xmember1"}' | jq

删除
curl -s -X DELETE http://localhost:8090/multisigs/0xabc123/members/0xmember1 | jq
```

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

## License

MIT License.
