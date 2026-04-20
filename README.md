# 民意智感中心 - 后端服务 (Golang)

## 项目概述

本项目是原 Django 项目 `PublicOpinionTexture` 的 Golang 重构版后端。

- 框架: Gin + GORM v2
- 数据库: MySQL (新库 `letter_manage_db`)
- 端口: 18081
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
  host: "81.70.230.137"
  port: 53306
  user: "your_db_user"        # 修改为实际用户
  password: "your_password"   # 修改为实际密码
  name: "letter_manage_db"
  charset: "utf8mb4"

llm:
  deepseek_api_key: "your_api_key"   # 修改为实际 DeepSeek API Key
  deepseek_api_url: "https://api.deepseek.com/chat/completions"
  deepseek_model: "deepseek-chat"

server:
  port: 18081
```

### 5. 启动服务

```bash
cd /Users/v_liheng02/work/test_code/for_test/letter-manage-frontend

# 方式1：直接运行
go run main.go

# 方式2：先构建再运行
go build -o letter-manage-server .
./letter-manage-server
```

服务启动后监听 http://localhost:18081

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
