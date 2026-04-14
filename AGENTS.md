# AGENTS.md · 项目上下文文档

> **⚠️ 重要：所有 AI 大模型在进入本项目时必须首先阅读本文件。**
> 本文件是项目的「单一事实来源」，描述了项目概况、技术选型、代码规范和禁止行为。

---

## 一、项目概况

这是一个面向个人用户的**全栈任务管理系统（TodoTask）**，采用 Monorepo 结构，使用 pnpm workspaces 管理多包依赖。

| 维度 | 说明 |
|------|------|
| 项目类型 | 全栈 Web 应用（个人任务管理） |
| 架构模式 | Monorepo（pnpm workspaces） |
| 包管理器 | **pnpm**（禁止 npm / yarn） |
| Node 版本 | `v20.10.0`（通过 nvm 锁定，根目录已配置 `.nvmrc`） |
| 核心功能 | 用户注册/登录（JWT）、任务 CRUD、多语言、主题色 |
| 工具脚本 | Node.js 脚本实现 MongoDB **数据自动化备份与恢复** |

> ⚠️ **大纲约束**：所有设计与开发必须以 [`docs/baseline.md`](./docs/baseline.md) 为最高权威，不得擅自偏离。

---

## 二、目录结构

```
todotask/
├── AGENTS.md                    ← 你正在读的文件（AI 项目上下文）
├── .nvmrc                       ← Node.js 版本锁定（v20.10.0）
├── package.json                 ← Monorepo 根配置（pnpm workspaces）
├── pnpm-workspace.yaml          ← workspace 声明
├── scripts/                     ← Node.js 工具脚本【待创建】
│   ├── backup.js                ← MongoDB 数据自动化备份
│   └── restore.js               ← MongoDB 数据恢复
├── docker-compose.yml           ← 本地开发容器编排【待创建】
├── .gitignore
│
├── docs/                        ← 项目文档（设计规范 & 进度）
│   ├── baseline.md              ← 项目大纲（最高权威，所有设计以此为准）
│   ├── dev-phases.md            ← 开发阶段计划 & 进度 checkbox
│   ├── api.md                   ← RESTful 接口文档
│   ├── backend-crud.md          ← 后端 CRUD 开发规范
│   ├── database-design.md       ← MongoDB 数据库设计
│   ├── frontend-prd.md          ← 前端 PRD 文档
│   └── testing.md               ← 自动化测试文档
│
├── agent/                       ← AI 编码规范（Skill 文件）
│   ├── backend/skill.md         ← Go 后端开发规范
│   ├── database/skill.md        ← MongoDB 数据库操作规范
│   ├── frontend/skill.md        ← Vue3 前端开发规范
│   ├── frontend/ui-skill.md     ← UI/UX & 动画规范
│   └── git/skill.md             ← Git 提交规范
│
└── packages/                    ← 业务代码【Phase 0 已完成】
    ├── backend/                 ← Go 后端服务【待初始化】
    │   ├── cmd/server/main.go
    │   ├── internal/handler/
    │   ├── internal/service/
    │   ├── internal/repository/
    │   ├── internal/model/
    │   ├── internal/middleware/
    │   ├── pkg/config/
    │   ├── pkg/logger/
    │   ├── pkg/response/
    │   └── configs/config.yaml
    └── frontend/                ← Vue3 前端应用【待初始化】
        ├── src/
        ├── index.html
        └── vite.config.ts
```

> ⚠️ **当前阶段**：Phase 0（工程初始化）尚未完成，`packages/` 目录不存在。
> 详细进度见 [`docs/dev-phases.md`](./docs/dev-phases.md)

---

## 三、技术栈速查

### 后端（`packages/backend/`）

| 类别 | 选型 |
|------|------|
| 语言 | Go 1.22+ |
| Web 框架 | Gin v1.9+ |
| 数据库 | MongoDB（官方 mongo-driver，禁止 ODM） |
| 配置管理 | Viper（读取 `configs/config.yaml` + 环境变量） |
| 日志 | Zap（结构化日志，**禁止 `fmt` 打日志**） |
| 认证 | JWT 双 Token（access 15min + refresh 7d 黑名单） |
| 容器化 | Docker + docker-compose |
| API 文档 | Apifox（Swagger 注解自动生成） |

### 前端（`packages/frontend/`）

| 类别 | 选型 |
|------|------|
| 框架 | Vue 3（`<script setup lang="ts">`，**禁止 Options API**） |
| 语言 | TypeScript（**禁止 `.js` 业务文件**） |
| 样式 | Tailwind CSS 为主，Less 处理复杂场景 |
| 状态管理 | Pinia + `localStorage` 持久化 |
| 多语言 | `vue-i18n`（中英双语） |
| UI 规范 | Tech-Noir 沉浸式科技风（GSAP 动画 + 霓虹色彩） |
| Lint | ESLint + Prettier |

---

## 四、后端架构分层

```
请求 → Handler → Service → Repository → MongoDB
```

| 层级 | 职责 | 路径 |
|------|------|------|
| Handler | 解析请求参数、映射错误到 HTTP 状态码、调用 Service | `internal/handler/` |
| Service | 业务逻辑编排、定义业务错误，调用 Repository | `internal/service/` |
| Repository | 只负责 MongoDB CRUD，将 DB 错误转为业务错误 | `internal/repository/` |
| Model | struct 定义，必须同时有 `bson` 和 `json` tag | `internal/model/` |
| Middleware | JWT 鉴权、CORS、限流、日志 | `internal/middleware/` |

**依赖方向单向**：Handler → Service → Repository，**禁止反向**。

---

## 五、统一 HTTP 响应结构

所有接口必须通过 `pkg/response` 包返回，**禁止直接调用 `c.JSON`**：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## 六、数据库关键规范

- **集合命名**：全小写复数，下划线分隔，如 `users`、`blog_posts`、`token_blacklist`
- **所有 DB 操作**必须携带 `context.Context` 超时控制（5s）
- **禁止物理删除**：统一软删除（`is_deleted: true` + `deleted_at`）
- **分页**：超过 10 页或 10 万条数据必须用游标分页，**禁止 `$skip` 深分页**
- **聚合管道**：`$match` 必须在首位，已验证索引才能上线

---

## 七、前端关键规范

- **组件文件名**：PascalCase（`UserCard.vue`），模板中使用 `<UserCard />`
- **命名**：composable 用 `use` 前缀（`useUserStore`），事件处理用 `handle` 开头
- **样式**：组件私有样式必须加 `scoped`，穿透子组件用 `:deep()`
- **import 顺序**：Vue 核心 → 第三方 → stores → composables → 组件 → types/utils/api
- **路由懒加载**：`() => import('@/views/...')`
- **禁止**：`any` 类型、index 作为 `v-for` key、`v-if` 与 `v-for` 同级
- **风格**：优雅的 UI/UX 风格，严谨的产品逻辑（参见 [`agent/frontend/ui-skill.md`](./agent/frontend/ui-skill.md)）

---

## 八、Git 提交规范

格式：`<type>(<scope>): <说明>`（说明**中文，不超过 25 字**）

| type | 含义 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 重构 |
| `style` | 代码格式调整 |
| `docs` | 文档更新 |
| `chore` | 工程配置 |
| `ci` | CI/CD 配置 |
| `build` | 构建/Docker 变更 |
| `revert` | 回滚 |

**示例**：`feat(auth): 新增 JWT 双 Token 登录接口`

> ⚠️ 未收到用户明确指令，**禁止执行 `git push`**，只做本地 commit。

---

## 九、AI 行为准则

在修改任何代码前，**必须**先加载对应的 Skill 文件：

| 修改范围 | 必读 Skill |
|----------|------------|
| `packages/backend/` 任意 Go 文件 | [`agent/backend/skill.md`](./agent/backend/skill.md) |
| `packages/backend/internal/repository/` 或 `model/` | [`agent/database/skill.md`](./agent/database/skill.md) |
| `packages/frontend/` 任意文件 | [`agent/frontend/skill.md`](./agent/frontend/skill.md) |
| 涉及首页/沉浸式 UI 页面 | [`agent/frontend/ui-skill.md`](./agent/frontend/ui-skill.md) |
| 执行 `git commit` | [`agent/git/skill.md`](./agent/git/skill.md) |

---

## 十、项目文档索引与使用时机

> **AI 开始任何开发任务前，必须先查阅对应文档，确认当前阶段和设计规范，不得擅自发挥。**

### 文档速查表

| 文档 | 路径 | 何时使用 |
|------|------|----------|
| **项目大纲** | [`docs/baseline.md`](./docs/baseline.md) | **最高权威**，所有设计不得偏离大纲，有疑问时以此为准 |
| 开发阶段计划 | [`docs/dev-phases.md`](./docs/dev-phases.md) | **每次开始新任务前必读**，确认当前处于哪个 Phase，对应 checkbox 完成后及时打勾 |
| 接口文档 | [`docs/api.md`](./docs/api.md) | 开发后端接口或前端对接 API 时，以此为准，不得擅自新增或改变接口结构 |
| 后端 CRUD 规范 | [`docs/backend-crud.md`](./docs/backend-crud.md) | 实现任何增删改查接口时，对照分层流程和错误处理约定 |
| 数据库设计文档 | [`docs/database-design.md`](./docs/database-design.md) | 编写 model/repository 代码时，以此为准定义字段和索引 |
| 前端 PRD 文档 | [`docs/frontend-prd.md`](./docs/frontend-prd.md) | 开发所有前端页面和组件前，先确认页面清单和功能需求 |
| 自动化测试文档 | [`docs/testing.md`](./docs/testing.md) | Phase 5 创建测试脚本时遵循，测试用例以文档清单为准 |

### AI 开发工作流（每次任务必须遵循）

```
1. 读 docs/dev-phases.md → 确认当前 Phase 和待完成的 checkbox
2. 读对应功能文档（api.md / backend-crud.md / frontend-prd.md）
3. 读对应 Skill 文件（agent/backend/skill.md 等）
4. 编写代码，完成后更新 dev-phases.md 中对应 checkbox
5. 执行 git commit（遵循 agent/git/skill.md）
```


---

## 十一、全局禁止事项（最高优先级）

| # | 禁止行为 |
|---|----------|
| 1 | 使用 `npm` 或 `yarn`（pnpm only） |
| 2 | 前端使用 Options API |
| 3 | 前端使用 `any` 类型 |
| 4 | 后端忽略 `error` 返回值（`result, _ := ...`） |
| 5 | 后端直接调用 `c.JSON` 返回响应 |
| 6 | 后端用 `fmt.Println` 打日志（必须用 Zap） |
| 7 | 数据库物理删除业务数据 |
| 8 | 未经明确授权执行 `git push` |
| 9 | 在 Handler 层写 MongoDB 查询 |
| 10 | 启动无法退出的 goroutine（fire-and-forget） |

