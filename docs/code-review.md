# 代码审查规范（提交前必做）

> 目标：每次提交代码前，对“本次待提交变更”执行统一代码审查，确保质量可控、可维护、可协作。
> 适用范围：`packages/frontend/`、`packages/backend/` 及相关文档变更。

---

## 一、审查原则

- 仅审查本次提交相关变更（`git diff --staged` 对应范围）。
- 先保证正确性，再追求优雅与性能。
- 审查意见必须可执行，避免空泛描述。
- 结论分级：`阻塞`（必须修复）/ `建议`（可在后续迭代）。

---

## 二、八维度审查清单（强制）

### 1. Design（设计）

- 代码设计是否符合当前系统分层与边界？
- 后端是否遵循 `Handler -> Service -> Repository` 单向依赖？
- 前端是否遵循页面、组件、store、api 的职责划分？
- 是否出现跨层调用、重复实现或隐式耦合？

### 2. Complexity（复杂性）

- 实现是否可以更简单？
- 是否存在过度抽象、过深嵌套、难理解的流程分支？
- 新同学是否能在短时间内读懂并安全修改？

### 3. Naming（命名）

- 变量、函数、类型、组件命名是否清晰表达意图？
- 命名是否符合项目规范（前端 `use*` / `handle*`，后端 `Err*` 等）？
- 是否存在歧义命名（如 `data`, `temp`, `obj`）？

### 4. Functionality（功能性）

- 行为是否符合需求意图与用户收益？
- 是否考虑边界条件、异常路径、权限与数据归属？
- 是否引入行为回归（特别是鉴权、分页、软删除、日志）？

### 5. Tests（测试）

- 是否补充/更新了自动化测试（至少覆盖关键路径）？
- 是否验证了失败路径和边界条件？
- 当前阶段无完整自动化时，是否至少附带可复现的手工验证步骤？

### 6. Comments（注释）

- 注释是否解释“为什么”，而不仅是“做了什么”？
- 复杂逻辑是否有必要注释辅助理解？
- 是否存在过期、误导或无意义注释？

### 7. Style（风格）

- 是否遵循项目规范：`AGENTS.md` + 对应 `agent/*/skill.md`？
- 前端是否通过 ESLint/Prettier/TypeScript 校验？
- 后端是否通过 `go vet` 与 `go test` 基础校验？
- 是否存在无意义调试代码（如 `console.log` / `fmt.Println`）？

### 8. Documentation（文档）

- 是否同步更新受影响文档（`docs/api.md`、`docs/frontend-prd.md`、`docs/testing.md` 等）？
- 对外行为变化（接口、配置、运行方式）是否写入文档？

---

## 三、提交前执行命令（强制）

在仓库根目录执行：

```bash
pnpm review:precommit
```

等价分步命令：

```bash
pnpm check:frontend
pnpm check:backend
```

判定标准：

- 任一命令失败：禁止提交，先修复再复审。
- 全部通过：进入提交环节。

---

## 四、提交说明要求

提交信息保持：

```text
<type>(<scope>): <中文说明>
```

- 说明不超过 25 字。
- 仅描述本次核心变更。
- 涉及规范类改动时，说明中需明确“审查/校验”相关关键词。

---

## 五、审查结果记录模板

每次提交前建议按以下模板记录（可放在 PR 描述或提交备注中）：

```text
[Code Review]
- Design: 通过 / 问题点
- Complexity: 通过 / 问题点
- Naming: 通过 / 问题点
- Functionality: 通过 / 问题点
- Tests: 通过 / 问题点
- Comments: 通过 / 问题点
- Style: 通过 / 问题点
- Documentation: 通过 / 问题点

[Validation]
- pnpm check:frontend: 通过/失败
- pnpm check:backend: 通过/失败
```

---

## 六、与现有规范的关系

- 本文档为“提交前审查流程规范”。
- 具体编码规范仍以以下文档为准：
  - `AGENTS.md`
  - `agent/frontend/skill.md`
  - `agent/backend/skill.md`
  - `docs/baseline.md`

