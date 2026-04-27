<div align="center">

# TodoTask

**沉浸式个人任务管理系统**

[![Node](https://img.shields.io/badge/Node-v20.10.0-brightgreen?logo=node.js)](https://nodejs.org)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org)
[![MongoDB](https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb)](https://mongodb.com)
[![pnpm](https://img.shields.io/badge/pnpm-workspace-F69220?logo=pnpm)](https://pnpm.io)
[![License](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

[English](./README.en.md) · 简体中文

</div>

---

## ✨ 项目介绍

TodoTask 是一个面向个人用户的**全栈任务管理系统**，采用 Tech-Noir 沉浸式科技风 UI 设计，支持任务 CRUD、JWT 认证、中英双语切换和多种霓虹主题色。

## 🧭 入口规则

- 首页 `/` 是统一入口，重新打开浏览器或访问根路径时优先展示首页。
- 已登录用户：首页仅展示「立即开始」，点击进入任务列表 `/tasks`。
- 未登录用户：首页仅展示「去登录」，点击进入 `/login`。
- 登录成功后直接进入任务列表；注册成功后回到登录页。

## 🖥️ 技术栈

| 端 | 技术 |
|----|------|
| **前端** | Vue 3 · TypeScript · Tailwind CSS · Pinia · vue-i18n · Vue Particles (tsparticles) |
| **后端** | Go 1.22 · Gin · JWT 双 Token |
| **数据库** | MongoDB 7.0（mongo-driver/v2） |
| **工程** | Monorepo · pnpm workspaces · Docker · docker-compose |

## 📁 目录结构

```
todotask/
├── .nvmrc                    # Node 版本锁定（v20.10.0）
├── docker-compose.yml        # 本地容器编排
├── auth/                     # 本地敏感信息（已 gitignore，不提交）
├── packages/
│   ├── backend/              # Go 后端服务
│   │   ├── cmd/server/       # 程序入口
│   │   ├── internal/         # 业务层（handler / service / repository）
│   │   ├── pkg/              # 公共工具（config / logger / response）
│   │   └── configs/          # 应用配置
│   └── frontend/             # Vue3 前端应用
│       └── src/
│           ├── views/        # 页面组件
│           ├── stores/       # Pinia 状态管理
│           ├── router/       # 路由（懒加载）
│           └── styles/       # 全局样式 & Less 变量
├── docs/                     # 项目文档
├── agent/                    # AI 编码规范（Skill 文件）
├── mcp/                      # MCP 工具集（Apifox 后端/前端同步）
└── scripts/                  # 数据备份 & 恢复脚本
```

## 🚀 快速开始

### 环境依赖

- [Node.js v20.10.0](https://nodejs.org)（推荐通过 `nvm use` 切换）
- [pnpm](https://pnpm.io) >= 8.0
- [Go](https://golang.org) >= 1.22
- [Docker](https://docker.com) & docker-compose

### 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/zhangyiran6866-lgtm/todo-task.git
cd todo-task

# 2. 切换 Node 版本
nvm use

# 3. 安装依赖
pnpm install

# 4. 启动 MongoDB（Docker）
pnpm docker:up

# 5. 启动开发服务器
# 方案 A: 一键启动前后端（推荐）
pnpm dev

# 方案 B: 分别启动（适合调试）
# 终端 1: 启动后端
pnpm dev:backend
# 终端 2: 启动前端
pnpm dev:frontend
```

- 前端：http://localhost:5173
- 后端：http://localhost:8080

### 📦 脚本命令

| 命令 | 说明 |
|------|------|
| `pnpm dev` | **推荐**：同时启动前端与后端开发环境 |
| `pnpm dev:backend` | 仅启动后端服务 |
| `pnpm dev:frontend` | 仅启动前端开发服务器 |
| `pnpm build:backend` | 编译后端二进制文件 |
| `pnpm build:frontend` | 构建前端生产包 |
| `pnpm docker:up` | 启动 MongoDB 容器（后台运行） |
| `pnpm docker:down` | 停止所有容器 |
| `pnpm backup` | 备份 MongoDB 数据（含自动清理） |
| `pnpm restore` | 交互式恢复 MongoDB 数据 |

## 🎨 UI 主题

支持 4 种霓虹主题色，在个人设置页切换：

| 主题 | 颜色 |
|------|------|
| Cyan（默认） | `#00f3ff` |
| Purple | `#bc13fe` |
| Green | `#39ff14` |
| Pink | `#ff00e5` |

## 📖 开发文档

| 文档 | 说明 |
|------|------|
| [项目大纲](docs/baseline.md) | 最高权威，所有设计以此为准 |
| [开发阶段计划](docs/dev-phases.md) | Phase 划分与进度追踪 |
| [接口文档](docs/api.md) | RESTful API 参考 |
| [数据库设计](docs/database-design.md) | MongoDB 集合设计 |
| [前端 PRD](docs/frontend-prd.md) | 页面与功能规范 |
| [测试文档](docs/testing.md) | 自动化测试规范 |
| [Apifox Backend MCP](mcp/apifox-backend/README.md) | 后端接口文档同步到 Apifox |
| [Apifox Frontend MCP](mcp/apifox-frontend/README.md) | Apifox 接口契约同步到前端 API TS 层 |

## 🔐 本地敏感信息

本项目使用根目录 `auth/` 保存本机 token、账号、API Key 等敏感信息，例如 `auth/apifox.md`。该目录已加入 `.gitignore`，不要提交到远端仓库。

## 🤝 贡献指南

提交前请遵循 [Git 提交规范](agent/git/skill.md)：

```
<type>(<scope>): <中文说明，不超过25字>

示例：feat(auth): 新增 JWT 双 Token 登录接口
```

## 📄 License

[MIT](LICENSE)
