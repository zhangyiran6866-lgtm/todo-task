# 全栈项目开发任务书

### 1. 环境与规范

- **管理：** 使用 `nvm` 锁定 Node.js 版本 `v20.10.0`。
- **结构：** 推荐 **Monorepo** 架构（pnpm workspaces），共享类型定义。
- **脚本：** 编写 Node.js 脚本实现数据的 **自动化备份与恢复**。

### 2. 后端服务 (Golang + MongoDB)

- **核心：** 实现用户认证（JWT）及个人任务管理系统的 **CRUD**。
- **安全：** 所有接口集成认证拦截器。
- **部署：** 编写 Dockerfile，通过 `docker build/run/exec` 实现容器化调试。
- **文档：** 使用 **Apifox** 维护接口文档。

### 3. 前端应用 (Vue3 + TS + Less)

- **状态：** **Pinia** + `localStorage` 持久化，支持任务筛选。
- **样式：** Less 变量/嵌套实现主题色管理。
- **多语言：** `vue-i18n` 实现中英双语切换。
- **规范：** 严谨的 **TS 类型定义** 与 **ESLint** 约束。
- **风格：** 优雅的UIUX风格，严谨的产品逻辑，引入 `tsparticles` 等粒子特效实现 Tech-Noir 赛博霓虹沉浸视觉。

### 4. 编码指南参考

**Skill 规范**

- [后端开发规范](../agent/backend/skill.md)
- [数据库开发规范](../agent/database/skill.md)
- [前端开发规范](../agent/frontend/skill.md)
- [前端 UI 规范](../agent/frontend/ui-skill.md)
- [Git 规范](../agent/git/skill.md)

**项目文档**

- [数据库设计文档](./database-design.md)
- [接口文档（API Reference）](./api.md)
- [后端 CRUD 开发规范](./backend-crud.md)
- [前端 PRD 文档](./frontend-prd.md)
- [开发阶段计划](./dev-phases.md)
- [自动化测试文档](./testing.md)
