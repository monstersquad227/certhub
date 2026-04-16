# CertHub

**Languages:** [English](README.md) | [中文](README.zh.md)

## Introduction

CertHub is a web application for applying for and managing SSL/TLS certificates (including Let's Encrypt via ACME). Users sign in with email verification codes, request certificates for domains, view and download private keys; it includes balance and recharge records, plus an admin area for certificate metadata.

## Features

- Email one-time code login; first-time email creates an account
- ACME-based issuance and DNS validation
- User certificate list, detail, and private key download
- Balance and recharge order history
- Admin JWT-protected certificate CRUD

## Tech stack

| Layer | Technology |
|--------|------------|
| Frontend | Vue 3, Vite, TypeScript, Ant Design Vue, Pinia, Vue Router |
| Backend | Go 1.24+, Gin, GORM, MySQL, Viper |
| Other | JWT, SMTP (verification email) |

## Repository layout

```
certhub/
├── backend/          # Go API (config directory — see below)
├── frontend/         # Vue 3 + Vite
├── README.md         # Chinese README
└── README.en.md      # This file (English)
```

## Requirements

- Go 1.24+
- Node.js 18+ (LTS recommended)
- MySQL 5.7+ / 8.x

## Quick start

### 1. Database

Create an empty database; `utf8mb4` is recommended. On startup the app runs GORM auto-migrations.

### 2. Backend configuration

`backend/config/` is typically gitignored. Add `backend/config/config.yaml`. Set the default environment with top-level `env` (e.g. `prod`), or override with `CERHUB_ENV` (e.g. `dev`). Under `database`, define one block per environment name matching `env` / `CERHUB_ENV`:

```yaml
env: prod

server:
  port: "8080"
  mode: debug   # debug or release

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
  key: "0123456789abcdef0123456789abcdef"  # 16 / 24 / 32 bytes
```

From the `backend` directory:

```bash
go mod download
go run .
```

The API listens on `http://localhost:8080` by default, with routes under `/api/v1`.

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server defaults to `http://localhost:5173` and proxies `/api` to `http://localhost:8080`.

Production:

```bash
npm run build
```

Serve `frontend/dist` behind your static host and align API origin or CORS/reverse proxy.

## API overview (partial)

| Method | Path | Notes |
|--------|------|--------|
| POST | `/api/v1/auth/send-code` | Send login code |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/admin/auth/login` | Admin login |
| POST | `/api/v1/certificates/generate-dns` | DNS validation (user JWT) |
| POST | `/api/v1/certificates/generate` | Issue certificate (user JWT) |
| GET | `/api/v1/certificates` | User certificates |
| GET | `/api/v1/balance` | Balance |
| … | `/api/v1/admin/...` | Admin certificate APIs (admin JWT) |

See `backend/internal/routes/routes.go` and controllers for details.

## Contributing

Issues and pull requests are welcome. Please match existing style and describe motivation and scope.

## License

If you open-source this repository, add a `LICENSE` file at the root and update this section accordingly.
