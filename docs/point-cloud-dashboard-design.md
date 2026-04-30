# 3D 点云看板设计方案（Stanford Bunny）

## 1. 目标与范围

- 在 TodoTask 中新增一个受鉴权保护的子页面：`/point-cloud`
- 基于根目录 `data/bunny` 数据集渲染 3D 点云看板
- 支持 Shader 动态颜色切换
- 在任务系统顶部导航栏增加入口按钮
- 本阶段先打通页面入口与基础结构，后续分步接入渲染能力

## 2. 信息架构

- 入口位置：`TasksView` 顶部导航右侧按钮（与语言/主题/用户菜单同层）
- 路由：`/point-cloud`（`requiresAuth: true`）
- 页面结构：
  - 顶部工具栏：返回任务列表、模型切换、渲染参数面板入口
  - 中央画布区：WebGL Canvas（Three.js）
  - 底部状态栏：点数/FPS/当前模型/加载状态

## 3. 数据策略

- 数据目录：`/data/bunny`
- 首选默认模型：`data/bunny/reconstruction/bun_zipper.ply`
- 预留可选模型：
  - `bun_zipper_res2.ply`
  - `bun_zipper_res3.ply`
  - `bun_zipper_res4.ply`
  - `data/` 子目录下多视角文件
- 性能兜底：高点数模型自动降采样或建议切换低分辨率模型

## 4. 渲染方案

- 技术选型：`three` + `PLYLoader` + `OrbitControls`
- 渲染对象：`THREE.Points` + `BufferGeometry` + `ShaderMaterial`
- 首版 Shader 模式：
  - Neon Cyan：统一霓虹青
  - Height Gradient：按高度渐变
  - Normal Glow：按法线映射伪彩
  - Pulse：随时间流动颜色
- Uniform 规划：
  - `uTime`
  - `uPointSize`
  - `uColorA`
  - `uColorB`
  - `uIntensity`

## 5. 交互设计

- 鼠标交互：旋转/缩放/平移（OrbitControls）
- 视角预设：Front / Left / Top / Iso
- 一键重置：恢复初始相机和参数
- Shader 切换：300~500ms 平滑过渡

## 6. 性能与稳定性

- 目标性能：中等模型桌面端 40~60 FPS
- 优化策略：
  - 片元计算最小化，尽量放在顶点阶段
  - 页面不可见时暂停动画循环
  - 路由离开后释放 `geometry/material/controls/renderer`
- 兼容兜底：WebGL 不可用时提供降级提示

## 7. 分步实施计划

1. 第一步（当前）：新增路由、导航入口、页面骨架与 i18n 文案
2. 第二步：接入 Three.js 基础场景与单模型点云渲染
3. 第三步：接入 Shader 动态切换和参数面板
4. 第四步：加入模型切换、性能优化和异常兜底

## 8. 涉及文件（规划）

- `packages/frontend/src/router/index.ts`
- `packages/frontend/src/views/tasks/TasksView.vue`
- `packages/frontend/src/views/point-cloud/PointCloudView.vue`
- `packages/frontend/src/locales/zh.ts`
- `packages/frontend/src/locales/en.ts`
