<div align="center">

# TodoTask

**Immersive Personal Task Management System**

[![Node](https://img.shields.io/badge/Node-v20.10.0-brightgreen?logo=node.js)](https://nodejs.org)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org)
[![MongoDB](https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb)](https://mongodb.com)
[![pnpm](https://img.shields.io/badge/pnpm-workspace-F69220?logo=pnpm)](https://pnpm.io)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

[简体中文](./README.md) · English

</div>

---

## ✨ Overview

TodoTask is a **full-stack personal task management system** built with an immersive Tech-Noir UI. It features task CRUD operations, JWT authentication, bilingual (Chinese/English) support, and multiple neon color themes.

## 🧭 Entry Rules

- The home page `/` is the unified entry point and is shown first when reopening the browser or visiting the root path.
- Logged-in users only see "Get Started" on the home page; clicking it opens the task list at `/tasks`.
- Logged-out users only see "Go to Login" on the home page; clicking it opens `/login`.
- After a successful login, users go directly to the task list; after registration, users return to the login page.

## 🖥️ Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Vue 3 · TypeScript · Tailwind CSS · Pinia · vue-i18n · Vue Particles (tsparticles) |
| **Backend** | Go 1.22 · Gin · JWT Dual Token |
| **Database** | MongoDB 7.0 (mongo-driver/v2) |
| **Infra** | Monorepo · pnpm workspaces · Docker · docker-compose |

## 📁 Project Structure

```
todotask/
├── .nvmrc                    # Node version lock (v20.10.0)
├── docker-compose.yml        # Local container orchestration
├── auth/                     # Local secrets (gitignored, do not commit)
├── packages/
│   ├── backend/              # Go backend service
│   │   ├── cmd/server/       # Application entry point
│   │   ├── internal/         # Business layers (handler / service / repository)
│   │   ├── pkg/              # Shared utilities (config / logger / response)
│   │   └── configs/          # Application configuration
│   └── frontend/             # Vue3 frontend application
│       └── src/
│           ├── views/        # Page components
│           ├── stores/       # Pinia state management
│           ├── router/       # Routes (lazy-loaded)
│           └── styles/       # Global styles & Less variables
├── docs/                     # Project documentation
├── agent/                    # AI coding standards (Skill files)
├── mcp/                      # MCP tools (e.g., apifox-backend)
└── scripts/                  # Data backup & restore scripts (Phase 5 pending)
```

## 🚀 Getting Started

### Prerequisites

- [Node.js v20.10.0](https://nodejs.org) (recommend using `nvm use`)
- [pnpm](https://pnpm.io) >= 8.0
- [Go](https://golang.org) >= 1.22
- [Docker](https://docker.com) & docker-compose

### Local Development

```bash
# 1. Clone the repository
git clone https://github.com/zhangyiran6866-lgtm/todo-task.git
cd todo-task

# 2. Switch Node version
nvm use

# 3. Install frontend dependencies
pnpm install

# 4. Start MongoDB (Docker)
docker-compose up -d mongodb

# 5. Start backend
cd packages/backend
go run ./cmd/server/main.go

# 6. Start frontend (new terminal)
pnpm dev:frontend
```

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Health Check: http://localhost:8080/health

### Docker Deployment

```bash
docker-compose up -d
```

## 📦 Scripts

| Command | Description |
|---------|-------------|
| `pnpm dev:frontend` | Start frontend dev server |
| `pnpm build:frontend` | Build frontend for production |
| `pnpm docker:up` | Start all containers |
| `pnpm docker:down` | Stop all containers |
| `pnpm backup` | Backup MongoDB data (Phase 5 pending) |
| `pnpm restore` | Restore MongoDB data (Phase 5 pending) |

> Note: `scripts/backup.js` and `scripts/restore.js` have not been created yet. The backup/restore commands will be available after Phase 5 is completed.

## 🎨 Themes

4 neon color themes available in the profile page:

| Theme | Color |
|-------|-------|
| Cyan (default) | `#00f3ff` |
| Purple | `#bc13fe` |
| Green | `#39ff14` |
| Pink | `#ff00e5` |

## 📖 Documentation

| Doc | Description |
|-----|-------------|
| [Baseline](docs/baseline.md) | Project charter — highest authority |
| [Dev Phases](docs/dev-phases.md) | Phase breakdown & progress tracking |
| [API Reference](docs/api.md) | RESTful API documentation |
| [Database Design](docs/database-design.md) | MongoDB schema design |
| [Frontend PRD](docs/frontend-prd.md) | Page & feature specifications |
| [Testing Guide](docs/testing.md) | Automated testing standards |
| [Apifox MCP](mcp/apifox-backend/README.md) | Backend API documentation sync tool |

## 🔐 Local Secrets

Use the root `auth/` directory for local-only tokens, accounts, API keys, and similar secrets, such as `auth/apifox.md`. This directory is ignored by `.gitignore`; do not commit it to the remote repository.

## 🤝 Contributing

Please follow the [Git Commit Standards](agent/git/skill.md) before making commits:

```
<type>(<scope>): <description in Chinese, max 25 chars>

Example: feat(auth): 新增 JWT 双 Token 登录接口
```

## 📄 License

[MIT](LICENSE)
