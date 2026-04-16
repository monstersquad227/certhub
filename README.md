# CertHub

<p align="center">
  <b>中文</b>
  &nbsp;·&nbsp;
  <a href="README.en.md">English</a>
</p>

---

## 简介

CertHub 是一套用于申请与管理 SSL/TLS 证书（含 Let's Encrypt ACME）的 Web 应用：用户通过邮箱验证码登录，按需为域名申请证书、查看与下载私钥；内置余额与充值流水；并提供管理端维护证书元数据。

## 特性

- 邮箱验证码登录；未注册邮箱首次登录即创建账号
- 基于 ACME（如 Let's Encrypt）的证书申请与 DNS 验证流程
- 用户侧证书列表、详情与私钥下载
- 余额与充值订单记录
- 管理端 JWT 鉴权下的证书 CRUD

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3、Vite、TypeScript、Ant Design Vue、Pinia、Vue Router |
| 后端 | Go 1.24+、Gin、GORM、MySQL、Viper |
| 其他 | JWT、SMTP（验证码邮件） |

## 仓库结构

```
certhub/
├── backend/          # Go API（配置目录见下文）
├── frontend/         # Vue 3 + Vite
├── README.md         # 本说明（中文）
└── README.en.md      # English README
```

## 环境要求

- Go 1.24+
- Node.js 18+（建议使用 LTS）
- MySQL 5.7+ / 8.x

## 快速开始

### 1. 数据库

创建空数据库，字符集建议使用 `utf8mb4`。应用启动时会执行 GORM 自动迁移并创建所需表。

### 2. 后端配置

仓库中 `backend/config/` 默认被忽略，请在 `backend/config/` 下新建 `config.yaml`。可通过根字段 `env` 指定默认环境（如 `prod`），或通过环境变量 `CERHUB_ENV` 覆盖（例如 `dev`）。`database` 下按环境名分多组，与 `env` / `CERHUB_ENV` 对应：

```yaml
env: prod

server:
  port: "8080"
  mode: debug   # debug 或 release

database:
  prod:
    host: 127.0.0.1
    port: 3306
    user: your_user
    password: your_password
    name: certhub
    charset: utf8mb4
  # dev:
  #   host: 127.0.0.1
  #   ...

jwt:
  secret: "change-me"
  expire_hours: 72

email:
  smtp_host: smtp.example.com
  smtp_port: 465
  username: ""
  password: ""
  from: "noreply@example.com"

security:
  verification_code_expire_minutes: 10
  verification_code_max_errors: 5
  verification_code_interval_seconds: 60

cert:
  price_single: 0
  price_wildcard: 0
  default_ca: letsencrypt
  storage_dir: ./data/certs

acme:
  directory_url: https://acme-v02.api.letsencrypt.org/directory
  email: your-acme-contact@example.com

aes:
  key: "0123456789abcdef0123456789abcdef"  # AES 密钥长度须为 16 / 24 / 32 字节
```

在 `backend` 目录执行：

```bash
go mod download
go run .
```

API 默认监听 `http://localhost:8080`，路由前缀为 `/api/v1`。

### 3. 前端

```bash
cd frontend
npm install
npm run dev
```

开发服务器默认 `http://localhost:5173`，并将 `/api` 代理到 `http://localhost:8080`。

生产构建：

```bash
npm run build
```

将 `frontend/dist` 交由静态服务器托管，并确保 API 同源或正确配置 CORS/反向代理。

## API 概要（节选）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/send-code` | 发送登录验证码 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/admin/auth/login` | 管理员登录 |
| POST | `/api/v1/certificates/generate-dns` | DNS 验证相关（需用户 JWT） |
| POST | `/api/v1/certificates/generate` | 申请证书（需用户 JWT） |
| GET | `/api/v1/certificates` | 用户证书列表 |
| GET | `/api/v1/balance` | 余额 |
| … | `/api/v1/admin/...` | 管理端证书管理（需管理员 JWT） |

具体行为以 `backend/internal/routes/routes.go` 与控制器实现为准。

## 参与贡献

欢迎 Issue 与 Pull Request。提交前请保持变更与现有代码风格一致，并说明动机与影响范围。

## 许可证

若你计划将本仓库开源，请在仓库根目录添加 `LICENSE` 文件并在此 README 中更新许可证名称与说明。
