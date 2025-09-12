# squads-rest-api

🚀 一个基于 Golang + MySQL 的 RESTful API 服务，支持 Multisig / Vault / Member / Spend 四个资源的完整 CRUD 操作，支持分页、搜索、过滤、排序。


## 数据库表结构 (MySQL Schema)
```
-- ---------- Multisigs ----------
CREATE TABLE IF NOT EXISTS `multisigs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_multisigs_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Vaults ----------
CREATE TABLE IF NOT EXISTS `vaults` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `vault_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_vaults_vault_address` (`vault_address`),
  KEY `idx_vaults_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Members ----------
CREATE TABLE IF NOT EXISTS `members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `member_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_members_member_address` (`member_address`),
  KEY `idx_members_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------- Spends ----------
CREATE TABLE IF NOT EXISTS `spends` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `spend_address` VARCHAR(255) NOT NULL,
  `multisig_address` VARCHAR(255) NOT NULL,
  `name` VARCHAR(255),
  `description` TEXT,
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_spends_spend_address` (`spend_address`),
  KEY `idx_spends_multisig_address` (`multisig_address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

```

## 功能特性

- Multisig / Vault / Member / Spend 四个表的 CRUD API
- 通用查询参数：分页、搜索、排序
- 子资源接口：
  - /multisigs/{multisig_address}/vaults
  - /multisigs/{multisig_address}/members
  - /multisigs/{multisig_address}/spends
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

```
curl -s http://localhost:8090/health | jq

响应：
{
  "success": true,
  "message": "ok"
}
```


## API 文档

```
Multisigs

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

Spends

创建
curl -s -X POST http://localhost:8090/multisigs/0xabc123/spends \
  -H "Content-Type: application/json" \
  -d '{"spend_address":"0xspend1"}' | jq

查询列表
curl -s "http://localhost:8090/multisigs/0xabc123/spends" | jq

查询单个
curl -s http://localhost:8090/multisigs/0xabc123/spends/111 | jq

更新
curl -s -X PUT http://localhost:8090/multisigs/0xabc123/spends/111 \
  -H "Content-Type: application/json" \
  -d '{"spend_address":"0xspend1"}' | jq

删除
curl -s -X DELETE http://localhost:8090/multisigs/0xabc123/spends/111 | jq
```

Vaults 和 Members的接口和Spends保持一致。

## License

MIT License.
