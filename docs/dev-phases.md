# 开发阶段计划

> **原则**：循序渐进，每个阶段有明确的完成标准（DoD），阶段间不跨越
> **进度追踪**：AI 协助开发时，应先查看当前阶段状态，再开始工作

---

## 阶段总览

| 阶段 | 名称 | 重点 | 状态 |
|------|------|------|------|
| Phase 0 | 工程初始化 | 基础设施搭建 | ✅ 完成 |
| Phase 1 | 认证模块 | 注册/登录/JWT | ✅ 完成 |
| Phase 2 | 任务 CRUD | 核心业务功能 | ✅ 完成 |
| Phase 3 | 前端核心页面 | 任务列表 UI | ✅ 完成 |
| Phase 4 | 体验优化 | 多语言/主题/动画 | ✅ 完成 |
| Phase 5 | 测试与上线 | 自动化测试 + Docker + 可观测性 | 🟡 进行中 |

---

## Phase 0 · 工程初始化 ✅

**目标**：搭建可运行的 Monorepo 骨架

- [x] AGENTS.md + Skill 文件体系建立
- [x] docs/ 项目文档体系（api / crud / db / prd / testing / dev-phases）
- [x] `.nvmrc` Node.js 版本锁定（v20.10.0）
- [x] `package.json` + `pnpm-workspace.yaml` Monorepo 根配置
- [x] `packages/backend/` Go 项目骨架（Gin + Viper + Zap）
- [x] `packages/frontend/` Vue3 + TS + Tailwind 初始化
- [x] `docker-compose.yml` + MongoDB 连接配置
- [x] `.gitignore` 配置
- [x] 后端 `/health` 健康检查接口

**DoD**：`docker-compose up` 后前后端均可访问，后端 `/health` 返回 200

---

## Phase 1 · 认证模块 ✅

**目标**：完成用户注册、登录、JWT 双 Token 机制

### 后端
- [x] `model/user.go` User 结构体
- [x] `repository/user_repository.go` 增删改查
- [x] `service/auth_service.go` 注册、登录、刷新、退出逻辑
- [x] `handler/auth_handler.go` 4 个接口（参见 `docs/api.md`）
- [x] JWT 双 Token 中间件（access 15min + refresh 7d 黑名单）

### 前端
- [x] `LoginView.vue` 登录页面
- [x] `RegisterView.vue` 注册页面
- [x] `useAuthStore.ts` Token 持久化
- [x] 路由守卫（未登录跳转 `/login`）

**DoD**：可完成注册 → 登录 → 获取用户信息 → 退出完整流程

---

## Phase 2 · 任务 CRUD ✅

**目标**：完成任务的增删改查后端接口

### 后端
- [x] `model/task.go` Task 结构体
- [x] `repository/task_repository.go` CRUD + 游标分页
- [x] `service/task_service.go` 业务逻辑（含归属权校验）
- [x] `handler/task_handler.go` 5 个接口（参见 `docs/api.md`）

**DoD**：通过 Apifox 可完整测试任务 CRUD 接口，含软删除验证

---

## Phase 3 · 前端核心页面 ✅

**目标**：完成任务列表 + 详情页，达到可用状态

- [x] `TasksView.vue` 任务列表主页面
- [x] `TaskCard.vue` 任务卡片组件
- [x] `TaskDetailView.vue` 任务详情/编辑页
- [x] `useTaskStore.ts` 任务状态管理
- [x] 滚动分页加载（游标分页）
- [x] 新建任务抽屉表单

**DoD**：用户可登录后完整执行任务 CRUD，UI 符合 Tech-Noir 风格

---

## Phase 4 · 体验优化 ✅

**目标**：完善多语言、主题色、动画细节

- [x] `vue-i18n` 中英双语接入
- [x] 主题色 CSS 变量切换（4 种霓虹色）
- [x] `ProfileView.vue` 个人信息页
- [x] 修改密码功能
- [x] 首页 Hero 动画精细化（GSAP）
- [x] 响应式布局适配（移动端）

**DoD**：切换语言/主题即时生效，移动端基本可用

---

## Phase 5 · 测试与上线 ⬜

**目标**：自动化测试覆盖核心接口 + 备份脚本 + Docker 生产部署 + 日志可观测性

### 测试
- [ ] Python 自动化测试脚本（参见 `docs/testing.md`）
- [ ] Go 单元测试（Service 层主要逻辑）

### 数据备份与恢复
- [x] `scripts/backup.js` — Node.js 脚本，调用 `mongodump` 备份数据库到本地参数化路径
- [x] `scripts/restore.js` — Node.js 脚本，调用 `mongorestore` 从指定备份总恢复
- [x] `package.json` 自动化脚本命令：`pnpm backup` / `pnpm restore`
- [x] 支持配置备份保留天数，自动清理过期备份

### 部署
- [ ] 生产环境 docker-compose 配置
- [ ] 环境变量 `.env` 管理
- [ ] README 更新

### 日志与可观测性
- [x] 后端统一结构化日志规范（JSON，含 `request_id`、`user_id`、`module`、`latency_ms`）
- [x] Gin 请求日志中间件（请求/响应摘要、错误码、耗时、来源 IP）
- [x] 日志分级与多通道输出（`app` / `error` / `audit`）
- [x] 日志持久化与保留策略（按天切割、压缩、过期清理）
- [x] 后端日志查询接口（按级别、关键词、时间范围、分页）
- [x] 前端日志中心入口（导航可达）与日志列表页面
- [x] 日志查看筛选能力（级别、时间范围、关键词）与详情抽屉
- [ ] 敏感字段脱敏策略（token/password/邮箱局部掩码）

**DoD**：自动化测试全部通过，`pnpm backup` 备份可执行，`pnpm restore` 恢复可验证，日志可在前端页面按条件检索并用于排障

---

## 进度更新规范

> 每次完成一个子项，将 `[ ]` 改为 `[x]`；阶段全部完成，更新阶段总览状态
> AI 开始新的开发任务前，**必须先确认当前处于哪个 Phase**
