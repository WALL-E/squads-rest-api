<div align="center">

# 🚀 Squads REST API

**一个高性能的多签钱包管理 REST API 服务**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=for-the-badge&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![API](https://img.shields.io/badge/API-REST-blue?style=for-the-badge)](docs/)

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [API 文档](#-api-文档) • [开发指南](#-开发指南)

</div>

---

## 📋 目录

- [功能特性](#-功能特性)
- [技术栈](#-技术栈)
- [快速开始](#-快速开始)
- [项目结构](#-项目结构)
- [API 文档](#-api-文档)
- [开发指南](#-开发指南)
- [部署](#-部署)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

## ✨ 功能特性

### 🔐 核心功能
- **多签钱包管理** - 完整的 Multisig 生命周期管理
- **金库系统** - Vault 资产管理和权限控制
- **成员管理** - 灵活的成员权限和角色分配
- **支出追踪** - 详细的 Spend 记录和审计

### 🛠 技术特性
- **RESTful API** - 标准化的 HTTP 接口设计
- **分页查询** - 高效的大数据集处理
- **搜索过滤** - 灵活的数据检索能力
- **排序支持** - 多字段排序功能
- **健康检查** - 服务状态监控
- **Swagger 文档** - 完整的 API 文档

### 🔗 子资源接口
```
/multisigs/{multisig_address}/vaults   # 多签钱包的金库
/multisigs/{multisig_address}/members  # 多签钱包的成员
/multisigs/{multisig_address}/spends   # 多签钱包的支出
```

## 🛠 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.21+ | 后端开发语言 |
| **Gin** | Latest | Web 框架 |
| **MySQL** | 8.0+ | 数据库 |
| **Swagger** | Latest | API 文档 |
| **Docker** | Latest | 容器化部署 |

## 🚀 快速开始

### 📦 安装要求

- **Go** 1.21 或更高版本
- **MySQL** 8.0 或更高版本
- **Git** (用于克隆代码)

### ⚡ 一键启动

```bash
# 1. 克隆项目
git clone https://github.com/yourname/squads-rest-api.git
cd squads-rest-api

# 2. 查看所有可用命令
make help

# 3. 设置数据库
make setup

# 4. 启动服务
make run
```

🎉 **服务启动成功！** 访问 http://localhost:8080

### 🔍 健康检查

```bash
curl -s http://localhost:8080/health | jq
```

**响应示例：**
```json
{
  "success": true,
  "message": "ok"
}
```

## 📁 项目结构

```
squads-rest-api/
├── 📁 cmd/                    # 应用程序入口
│   ├── 🖥️  server/main.go      # API 服务器
│   ├── 🧪 test/main.go        # API 测试工具
│   └── ⚙️  setup/main.go       # 数据库设置工具
├── 📁 docs/                   # API 文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── 📄 Makefile               # 构建脚本
├── 📄 README.md              # 项目文档
├── 📄 go.mod                 # Go 模块定义
└── 📄 go.sum                 # 依赖版本锁定
```


## 📚 API 文档

### 🌐 基础信息

- **Base URL:** `http://localhost:8080`
- **Content-Type:** `application/json`
- **响应格式:** JSON
- **Swagger 文档:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### 📋 API 概览

| 资源 | 端点 | 描述 |
|------|------|------|
| 🔐 **Multisigs** | `/multisigs` | 多签钱包管理 |
| 🏦 **Vaults** | `/multisigs/{multisig_address}/vaults` | 金库管理 |
| 👥 **Members** | `/multisigs/{multisig_address}/members` | 成员管理 |
| 💰 **Spends** | `/multisigs/{multisig_address}/spends` | 支出记录 |

---

<details>
<summary><strong>🔐 Multisigs API</strong></summary>

#### 创建多签钱包
```bash
curl -X POST http://localhost:8080/multisigs \
  -H "Content-Type: application/json" \
  -d '{
    "multisig_address": "0xabc123",
    "name": "Squad A",
    "description": "First squad"
  }'
```

#### 查询多签钱包列表
```bash
curl "http://localhost:8080/multisigs?page=1&limit=10&search=Squad"
```

#### 查询单个多签钱包
```bash
curl "http://localhost:8080/multisigs/0xabc123"
```

#### 更新多签钱包
```bash
curl -X PUT http://localhost:8080/multisigs/0xabc123 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Squad A Updated",
    "description": "Updated description"
  }'
```

#### 删除多签钱包
```bash
curl -X DELETE http://localhost:8080/multisigs/0xabc123
```

</details>

<details>
<summary><strong>💰 Spends API</strong></summary>

#### 创建支出记录
```bash
curl -X POST http://localhost:8080/multisigs/0xabc123/spends \
  -H "Content-Type: application/json" \
  -d '{
    "spend_address": "0xspend1",
    "amount": "1000000",
    "description": "Team payment"
  }'
```

#### 查询支出列表
```bash
curl "http://localhost:8080/multisigs/0xabc123/spends?page=1&limit=10"
```

#### 查询单个支出
```bash
curl "http://localhost:8080/multisigs/0xabc123/spends/0xspend1"
```

#### 更新支出记录
```bash
curl -X PUT http://localhost:8080/multisigs/0xabc123/spends/0xspend1 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated payment description"
  }'
```

#### 删除支出记录
```bash
curl -X DELETE http://localhost:8080/multisigs/0xabc123/spends/0xspend1
```

</details>

<details>
<summary><strong>🏦 Vaults API</strong></summary>

#### 创建金库
```bash
curl -X POST http://localhost:8080/multisigs/0xabc123/vaults \
  -H "Content-Type: application/json" \
  -d '{
    "vault_address": "0xvault1",
    "name": "Main Vault",
    "description": "Primary asset storage"
  }'
```

#### 查询金库列表
```bash
curl "http://localhost:8080/multisigs/0xabc123/vaults?page=1&limit=10"
```

#### 查询单个金库
```bash
curl "http://localhost:8080/multisigs/0xabc123/vaults/0xvault1"
```

#### 更新金库
```bash
curl -X PUT http://localhost:8080/multisigs/0xabc123/vaults/0xvault1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Vault Name",
    "description": "Updated description"
  }'
```

#### 删除金库
```bash
curl -X DELETE http://localhost:8080/multisigs/0xabc123/vaults/0xvault1
```

</details>

<details>
<summary><strong>👥 Members API</strong></summary>

#### 添加成员
```bash
curl -X POST http://localhost:8080/multisigs/0xabc123/members \
  -H "Content-Type: application/json" \
  -d '{
    "member_address": "0xmember1",
    "role": "admin",
    "permissions": ["read", "write", "execute"]
  }'
```

#### 查询成员列表
```bash
curl "http://localhost:8080/multisigs/0xabc123/members?page=1&limit=10"
```

#### 查询单个成员
```bash
curl "http://localhost:8080/multisigs/0xabc123/members/0xmember1"
```

#### 更新成员信息
```bash
curl -X PUT http://localhost:8080/multisigs/0xabc123/members/0xmember1 \
  -H "Content-Type: application/json" \
  -d '{
    "role": "member",
    "permissions": ["read"]
  }'
```

#### 移除成员
```bash
curl -X DELETE http://localhost:8080/multisigs/0xabc123/members/0xmember1
```

</details>

## 🛠 开发指南

### 📋 Makefile 命令

项目提供了完整的 Makefile 来管理不同的组件：

| 命令 | 描述 | 用途 |
|------|------|------|
| `make run` | 🚀 运行API服务器 | 开发调试 |
| `make test` | 🧪 运行API测试 | 功能测试 |
| `make setup` | ⚙️ 运行数据库设置 | 初始化环境 |
| `make build` | 🔨 构建所有组件 | 生产部署 |
| `make clean` | 🧹 清理构建文件 | 环境清理 |
| `make help` | ❓ 显示帮助信息 | 查看命令 |

### 🔧 开发环境配置

1. **环境变量配置**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=3306
   export DB_USER=root
   export DB_PASSWORD=password
   export DB_NAME=squads_db
   ```

2. **数据库初始化**
   ```bash
   make setup  # 自动创建数据库表结构
   ```

3. **开发模式启动**
   ```bash
   make run    # 启动开发服务器
   ```

### 🧪 测试

```bash
# 运行API测试套件
make test

# 手动测试健康检查
curl http://localhost:8080/health
```

## 🚀 部署

### 📦 生产构建

```bash
# 构建所有组件
make build

# 构建产物
ls -la server test-api setup-db
```

### 🐳 Docker 部署

```dockerfile
# Dockerfile 示例
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### ☁️ 云部署

支持部署到以下平台：
- **Docker** - 容器化部署
- **Kubernetes** - 集群部署
- **云服务器** - 直接部署

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 📝 贡献流程

1. **Fork** 本仓库
2. **创建** 特性分支 (`git checkout -b feature/AmazingFeature`)
3. **提交** 更改 (`git commit -m 'Add some AmazingFeature'`)
4. **推送** 到分支 (`git push origin feature/AmazingFeature`)
5. **创建** Pull Request

### 📋 开发规范

- 遵循 Go 代码规范
- 添加适当的测试用例
- 更新相关文档
- 确保所有测试通过

### 🐛 问题报告

发现 Bug？请 [创建 Issue](https://github.com/yourname/squads-rest-api/issues) 并包含：
- 详细的问题描述
- 复现步骤
- 期望行为
- 实际行为
- 环境信息

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给它一个星标！**

**🤝 欢迎贡献代码，让我们一起让它变得更好！**

[⬆ 回到顶部](#-squads-rest-api)

</div>
