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

## 🖥️ Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Vue 3 · TypeScript · Tailwind CSS · Pinia · vue-i18n · GSAP |
| **Backend** | Go 1.22 · Gin · JWT Dual Token |
| **Database** | MongoDB 7.0 (mongo-driver/v2) |
| **Infra** | Monorepo · pnpm workspaces · Docker · docker-compose |

## 📁 Project Structure

```
todotask/
├── .nvmrc                    # Node version lock (v20.10.0)
├── docker-compose.yml        # Local container orchestration
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
└── scripts/                  # Data backup & restore scripts
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
| `pnpm backup` | Backup MongoDB data |
| `pnpm restore` | Restore MongoDB data |

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

## 🤝 Contributing

Please follow the [Git Commit Standards](agent/git/skill.md) before making commits:

```
<type>(<scope>): <description in Chinese, max 25 chars>

Example: feat(auth): 新增 JWT 双 Token 登录接口
```

## 📄 License

[MIT](LICENSE)
