# 民意智感中心 - 后端服务 (Golang)

## 项目概述

本项目是原 Django 项目 `PublicOpinionTexture` 的 Golang 重构版后端。

- 框架: Gin + GORM v2
- 数据库: MySQL (新库 `letter_manage_db`)
- 端口: 8080（可通过 WORK_ENV 切换）
- 认证: Cookie + Session (session_key)

---

## 目录结构

```
letter-manage-frontend/        <- 注意：此目录名是后端代码目录
├── main.go                    # 入口，路由注册，CORS
├── go.mod                     # 模块 letter-manage-backend
├── config.yaml                # 配置文件（DB, LLM, Server等）
├── config/config.go           # 配置加载
├── middleware/auth.go         # 认证中间件
├── controller/                # API 控制器层
│   ├── auth.go                # POST /api/auth/
│   ├── letter.go              # POST /api/letter/
│   ├── setting.go             # POST /api/setting/
│   ├── config_api.go          # POST /api/config/
│   ├── llm.go                 # POST /api/llm/
│   └── tool.go                # POST /api/tool/
├── service/                   # 业务逻辑层
├── dao/                       # 数据访问层
├── model/                     # 数据模型
└── scripts/
    ├── init.sql               # 新库 DDL（建表脚本）
    ├── migrate.sql            # 数据迁移参考 SQL
    └── migrate_data.py        # Python 数据迁移脚本
```

---

## 快速开始

### 1. 环境要求

- Go 1.22+
- MySQL 5.7+ / 8.0
- 原数据库可访问 (只读迁移数据用)

### 2. 初始化新数据库

```bash
# 连接到 MySQL，执行建表脚本
mysql -h 81.70.230.137 -P 53306 -u <新用户> -p < scripts/init.sql
```

> 脚本会自动创建 `letter_manage_db` 数据库和所有表，并插入一个默认管理员账号（警号: admin, 密码: admin123）

### 3. 数据迁移（可选，将老数据导入新库）

```bash
# 安装依赖
pip3 install pymysql

# 编辑 scripts/migrate_data.py，修改 OLD_DB 和 NEW_DB 的连接信息
vim scripts/migrate_data.py

# 执行迁移（只读老库，不会修改原数据）
python3 scripts/migrate_data.py
```

### 4. 修改配置

编辑 `config.yaml`：

```yaml
database:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "000000"
  name: "letter_manage_db"

llm:
  api_key: "sk-..."   # 修改为实际 DeepSeek API Key
  api_url: "https://api.deepseek.com/v1/chat/completions"
  model: "deepseek-chat"

server:
  port: 8080
```

### 5. 启动服务

```bash
cd /Users/liheng/work/letter-manage-backend

# 方式1：直接运行
WORK_ENV=home go run main.go

# 方式2：先构建再运行
go build -o letter-manage-server main.go
WORK_ENV=home ./letter-manage-server
```

服务启动后监听 http://localhost:8080

---

## API 规范

所有接口均为 **POST**，请求体为 JSON：

```json
{
  "order": "操作命令",
  "args": { "参数key": "参数value" }
}
```

响应格式：

```json
{
  "success": true,
  "data": { },
  "error": ""
}
```

### 接口列表

| 路径 | 说明 |
|------|------|
| POST /api/auth/ | 认证（login/logout/check） |
| POST /api/letter/ | 信件管理（21种操作） |
| POST /api/config/ | 系统配置（菜单/系统信息） |
| POST /api/setting/ | 设置管理（分类/用户/单位/权限等） |
| POST /api/llm/ | AI 大模型（chat/chat_stream/get_prompt） |
| POST /api/tool/ | 工具接口（时间计算/工作日等） |

---

## 权限级别

| 级别 | 说明 | 可用功能 |
|------|------|----------|
| CITY | 市局 | 全部功能 |
| DISTRICT | 区县局/支队 | 下发/核查/用户管理 |
| OFFICER | 民警 | 信件查看/处理 |

---

## 信件状态流转

```
预处理
  ├─> 市局下发至区县局/支队  -> 已下发至处理单位 -> 正在处理
  │                                                  └─> 正在反馈 -> 待市局审核 -> 已办结
  └─> (越级) 已下发至处理单位 -> 正在处理
                                └─> 正在反馈 -> 待市局审核 -> 已办结
  └─> 已办结（标记无效）
```

---

## WORK_ENV 环境切换

系统支持通过 `WORK_ENV` 环境变量切换不同环境的配置，避免手动修改 `config.yaml`。

### 配置方式

```bash
# 使用 home 环境配置（本地开发）
export WORK_ENV=home
./letter-manage-server

# 使用 company 环境配置（公司/服务器）
export WORK_ENV=company
./letter-manage-server

# 不使用环境变量（使用 config.yaml 基础配置）
unset WORK_ENV
./letter-manage-server
```

### 支持的环境配置字段

并非所有字段都支持环境覆盖。目前支持覆盖的字段：

| 环境键 | 支持覆盖的字段 | 说明 |
|--------|---------------|------|
| `server` | `port`, `mode` | 端口和模式 |
| `database` | `host`, `port`, `user`, `password`, `name` | 数据库连接 |
| `llm` | `api_url`, `api_key`, `model`, `timeout` | AI 模型配置 |
| `media` | `root` | 媒体文件路径 |

### 当前配置的环境

在 `config.yaml` 的 `environments` 段定义：

```yaml
environments:
  home:
    database:
      host: 127.0.0.1
      user: root
      password: "000000"
      port: 3306
      name: letter_manage_db
    server:
      port: 8080
  company:
    database:
      host: 10.25.65.177
      user: root
      password: "000000"
      port: 8306
      name: letter_manage_db
```

> 启动时日志会打印 `WORK_ENV=home: applied environment overrides` 表示已成功应用环境覆盖。

---

## LLM_API_KEY 环境变量

DeepSeek API Key **只能通过** `LLM_API_KEY` 环境变量设置，配置文件（`config.yaml`）中不包含 api_key 字段。

### 配置方式

```bash
# 写入 shell profile（推荐）
echo 'export LLM_API_KEY="sk-your-key-here"' >> ~/.zshrc
source ~/.zshrc

# 或临时设置
export LLM_API_KEY="sk-your-key-here"
./letter-manage-server
```

> 启动时日志会打印 `LLM_API_KEY: applied environment override` 表示已生效。若未设置，AI 功能将不可用。

---

## 本地域名配置（多平台 Cookie 隔离）

当同时运行管理端（admin）和市民端（citizen）时，两个平台使用相同的 `session_key` cookie 名称。若部署在同一域名下，登录一个平台会覆盖另一个平台的 cookie，导致另一平台被迫下线。

### 解决方案

使用不同的本地域名访问两个平台，使 cookie 按域名隔离。

### 配置步骤

**1. 添加 hosts 映射**

```bash
sudo tee -a /etc/hosts << EOF

# letter-manage 平台本地域名（多平台 cookie 隔离）
127.0.0.1	admin.letter.local
127.0.0.1	citizen.letter.local
EOF
```

**2. 访问方式**

| 平台 | 本地地址 | Cookie 域 |
|------|---------|-----------|
| 管理端（admin） | http://admin.letter.local:5173 | `admin.letter.local` |
| 市民端（citizen） | http://citizen.letter.local:5174 | `citizen.letter.local` |
| 后端 API | http://localhost:8080 | （通过前端代理） |

**3. CORS 配置**

`config.yaml` 中的 `cors.allowed_origins` 已预置本地域名地址，如需添加更多域名，请在此处追加。
