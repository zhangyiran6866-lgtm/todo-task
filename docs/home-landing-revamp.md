# 首页入口重构定稿（2026-05-14）

## 背景

为提升首页导航效率与视觉一致性，调整首页入口结构：将任务管理、点云看板、图片标注三个功能直接放到首页首屏，替代原有通用文案指标项，并同步优化首屏视觉表现。

## 目标

- 首页成为统一功能入口层，减少二级页面跳转成本。
- 保持 Tech-Noir 暗色沉浸感，强化动态光效的识别度。
- 完成中英双语文案与首屏语言切换能力。

## 交互与信息架构

### 首页入口

首页主入口改为三个可点击胶囊按钮（从左到右）：

1. 任务管理系统
2. 点云看板
3. 图片标注

### 鉴权与跳转规则

- 任务管理系统：
  - 已登录：跳转 `/tasks`
  - 未登录：跳转 `/login`
- 点云看板：无需登录，跳转 `/point-cloud`
- 图片标注：无需登录，跳转 `/image-annotator`

### 路由访问控制

- `/tasks` 保持鉴权。
- `/point-cloud` 改为公开访问。
- `/image-annotator` 改为公开访问。

### 语言切换

- 首页首屏右上角新增独立悬浮语言切换图标。
- 支持中英双语切换，沿用现有 `themeStore.language` + `vue-i18n` 联动机制。

## 视觉与动效定稿

### 移除项

- 移除首页背景网格层（`hero-grid`）。
- 移除首页原主按钮区（“立即开始/去登录”）。
- 移除首页 `Focus / Plan / Execute` 指标文案。

### 光效方案

- 由双光晕升级为三光晕（RGB：红/绿/蓝）。
- 三个光晕在首屏内动态移动并发生碰撞反馈。
- 采用加色混合视觉：
  - 两两重叠可见青/洋红/黄倾向
  - 三色重叠中心为高饱和冷光核
- 明确约束：不出现“中心趋白”，整体保持暗色炫酷主题。

## 文案定稿（i18n）

- 中文主标题：`个人空间`
- 中文副标题：`永远相信美好即将发生`
- 英文主标题：`My Space`
- 英文副标题：`Believe that beautiful things are on the way.`

## 实施范围

- `packages/frontend/src/views/home/index.vue`
- `packages/frontend/src/router/index.ts`
- `packages/frontend/src/locales/zh.ts`
- `packages/frontend/src/locales/en.ts`

