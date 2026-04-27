# Phase 5: MongoDB 自动化备份与恢复实施方案

> **状态**：方案待评审
> **目标**：实现生产级别的数据库保障机制，确保数据可备份、可恢复、可清理。

---

## 一、 核心逻辑设计

采用 Node.js 封装 `mongodump` 与 `mongorestore` 工具，提供参数化、自动清理的备份机制。

### 1.1 自动备份脚本 (`scripts/backup.js`)
- **核心流程**：
    1.  **解析配置**：读取 `packages/backend/configs/config.yaml` 中的连接字符串（或优先读取环境变量）。
    2.  **执行 dump**：调用 `mongodump --uri="..." --out="backups/<TIMESTAMP>"`。
    3.  **自动压缩**：将备份目录压缩为 `.zip` 或 `.tar.gz` 格式以节省空间。
    4.  **清理逻辑**：扫描 `backups/` 目录，删除超过 `RETENTION_DAYS`（默认 7 天）的旧备份。
- **触发方式**：`pnpm backup`

### 1.2 安全恢复脚本 (`scripts/restore.js`)
- **核心流程**：
    1.  **备份自检**：列出 `backups/` 下所有可用备份供用户选择。
    2.  **安全确认**：要求用户手动输入 `CONFIRM`，防止在生产环境误操作覆盖。
    3.  **执行 restore**：调用 `mongorestore --drop --uri="..." <PATH_TO_BACKUP>`。
- **触发方式**：`pnpm restore`

---

## 二、 实施步骤计划

1.  **Step 1: 环境依赖校验**
    - 检查本地是否安装 `mongodump` 和 `mongorestore` 工具。
2.  **Step 2: 创建目录结构**
    - 创建根目录 `scripts/` 用于存放脚本。
    - 创建根目录 `backups/` 用于存放备份文件（已在 `.gitignore` 中配置排除）。
3.  **Step 3: 实现 `backup.js`**
    - 实现连接串解析 -> 执行 dump -> 自动清理过期文件的闭环逻辑。
4.  **Step 4: 实现 `restore.js`**
    - 实现交互式选择备份文件 -> 二次确认 -> 执行恢复的逻辑。
5.  **Step 5: 联调测试**
    - 手动造数 -> 备份 -> 删库 -> 恢复，验证 DoD。

---

## 三、 定义 DoD (完成标准)

- [ ] 执行 `pnpm backup` 后，在 `backups/` 目录下生成带时间戳的压缩文件。
- [ ] 连续执行 8 次备份（假设保留 7 次），最早的一份备份应被自动删除。
- [ ] 执行 `pnpm restore` 并确认后，数据库状态应完全恢复至所选备份点。

---
*方案起草人：OpenCode AI*
*日期：2026-04-27*

