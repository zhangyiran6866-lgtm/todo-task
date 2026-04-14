# Git 提交规范 · Skill

> **适用范围**：本项目所有 git commit 操作
> **权威来源**：Conventional Commits 规范 · 行业最佳实践 · 项目定制要求
> **强制级别**：AI 执行任何 `git commit` 操作前，必须先加载并严格遵守本文件。

---

## 一、核心原则

1. **中文优先**：提交说明统一使用中文，英文仅用于 type 前缀
2. **简短精炼**：提交说明（description）**不得超过 25 个字**
3. **原子提交**：每次提交只做一件事，功能与修复不混在同一条 commit
4. **禁止擅自推送**：未收到用户明确指令，只做本地 commit，绝不执行 `git push`

---

## 二、提交格式

```
<type>(<scope>): <description>
```

- **type**：变更类型（英文小写，见下表）
- **scope**：影响范围（可选，英文小写）
- **description**：简短说明（**中文，不超过 25 字**）

### 示例

```bash
feat(auth): 新增 JWT 双 Token 登录接口
fix(post): 修复文章列表分页错误
refactor(repo): 重构用户 Repository 层
style(home): 调整首页 Hero 区域间距
docs: 更新后端开发规范 Skill
chore: 升级 Gin 依赖至 v1.10
ci: 修复 GitHub Actions 构建失败
```

---

## 三、Type 类型定义

| Type | 含义 | 示例场景 |
|------|------|----------|
| `feat` | 新功能 | 新增接口、新增页面、新增组件 |
| `fix` | Bug 修复 | 修复逻辑错误、修复显示问题 |
| `refactor` | 重构（不改变功能） | 代码结构调整、分层优化 |
| `style` | 代码格式/样式调整 | 缩进、CSS 调整（无逻辑变更） |
| `perf` | 性能优化 | 优化查询、减少重渲染 |
| `test` | 测试相关 | 新增/修改单元测试 |
| `docs` | 文档变更 | 更新 README、Skill 文件、注释 |
| `chore` | 杂务/工程配置 | 依赖升级、配置文件修改 |
| `ci` | CI/CD 配置变更 | GitHub Actions workflow 修改 |
| `revert` | 回滚某次提交 | 撤销上一个有问题的 commit |
| `build` | 构建系统变更 | Dockerfile、docker-compose 修改 |

---

## 四、Scope 常用范围（可选）

| Scope | 含义 |
|-------|------|
| `auth` | 登录 / 鉴权模块 |
| `user` | 用户模块 |
| `post` | 博客文章模块 |
| `doc` | 文档/知识库模块 |
| `home` | 首页 |
| `admin` | 后台管理 |
| `handler` | 后端 handler 层 |
| `service` | 后端 service 层 |
| `repo` | 后端 repository 层 |
| `model` | 数据模型 |
| `middleware` | 中间件 |
| `config` | 配置相关 |
| `ci` | CI/CD 流程 |
| `deps` | 依赖管理 |

> Scope 不是强制的；影响面较广（如全局重构）可省略。

---

## 五、长度与格式规则

```
✅ feat(auth): 新增 refresh token 刷新接口        ← 17字，合规
✅ fix(post): 修复文章软删除未过滤问题             ← 15字，合规
✅ chore(deps): 升级 mongo-driver 至 v2           ← 16字，合规

❌ feat: Add user authentication with JWT double token refresh mechanism  ← 英文且过长
❌ fix: 修复了一个关于用户登录时 JWT access token 过期后无法通过 refresh token 续期的 bug  ← 超过25字
❌ update: 更新代码                                ← type 不合规，说明过于模糊
❌ feat+fix: 新增功能并修复 bug                    ← 一次 commit 做了两件事
```

---

## 六、破坏性变更（Breaking Change）

当某次变更会破坏已有接口或行为时，在 type 后加 `!` 标记：

```bash
feat(auth)!: JWT payload 结构调整，移除 username 字段
refactor(api)!: 响应结构统一改为 {code, message, data}
```

---

## 七、多文件提交原则（原子性）

```bash
# ✅ 正确：按逻辑拆分提交
git add internal/handler/user.go internal/service/user.go
git commit -m "feat(user): 新增用户信息查询接口"

git add internal/repository/user.go
git commit -m "feat(repo): 新增用户 Repository 查询实现"

# ❌ 错误：将无关变更堆在同一 commit
git add .
git commit -m "更新了一堆东西"
```

---

## 八、AI 执行提交的行为约束

1. **拆分提交**：如有多个逻辑独立的变更，必须分多次 commit，不得 `git add .` 一锅端
2. **说明内容**：提交前告知用户将要执行的 commit 内容，确认后再执行
3. **禁止 push**：未收到"推送"/"push"/"提交并推送"等明确指令，只执行本地 commit
4. **格式自检**：生成 commit message 后，自检是否符合格式规范和 25 字限制

### AI 提交前自检清单

```
□ type 是否来自规定列表？
□ description 是否为中文？
□ description 是否不超过 25 字？
□ 此次变更是否是单一逻辑（原子提交）？
□ 用户是否已明确授权执行 commit？
□ 用户是否明确要求 push？（否则只 commit）
```

---

## 九、禁止事项清单

| # | 禁止行为 | 原因 |
|---|----------|------|
| 1 | `git add .` 后直接提交所有变更 | 违反原子提交原则 |
| 2 | commit message 超过 25 字 | 违反简短规范 |
| 3 | 使用未定义的 type（如 `update`、`modify`、`change`） | 不规范，无法自动化处理 |
| 4 | 全英文 description | 本项目统一中文 |
| 5 | 未经用户确认执行 `git push` | 最高优先级约束 |
| 6 | 在一条 commit 中混合功能开发和 bug 修复 | 违反原子提交 |
| 7 | commit message 内容模糊（如"修改代码"、"更新"） | 无法追溯变更意图 |
