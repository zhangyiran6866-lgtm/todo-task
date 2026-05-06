# 图片标注功能实施计划（Paper.js）

## 1. 背景与目标

- 在 TodoTask 前端新增一个“图片标注”工具页面，支持用户上传图片并进行可视化标注。
- 标注图形覆盖常见简单几何形状：矩形、圆形/椭圆、三角形、五角星、心形。
- 标注元素支持样式与位置编辑：颜色、大小、描边线条、透明度、拖动。
- 支持将“原图 + 标注”导出为图片并下载到本地。

## 2. 范围定义

### 2.1 本期范围（MVP）

- 页面入口：在任务页顶部导航新增“图片标注”入口。
- 基础能力：
  - 上传图片（常见格式，如 png/jpg/webp）
  - 创建图形标注（矩形、圆、三角形、五角星、心形）
  - 选中与拖动标注元素
  - 调整填充色、描边色、线宽、虚线、透明度、整体尺寸
  - 删除选中元素
  - 导出标注结果为 PNG 并下载

### 2.2 暂不纳入本期

- 撤销/重做历史栈
- 多人协作标注
- 云端持久化与标注版本管理
- 高级路径编辑（锚点级编辑）

## 3. 方案总览

- 技术基座：Vue 3 + TypeScript + Paper.js
- 页面形态：独立路由页面 `/image-annotator`
- 分层结构：
  - 背景图层：用户上传图片（锁定，不可编辑）
  - 标注图层：所有图形元素（可选中、可拖拽、可样式编辑）
- 交互模型：
  - 工具模式：`select`（选择/拖动）与 `draw:<shape>`（创建图形）
  - 右侧属性面板根据当前选中元素动态展示可编辑属性

## 4. 关键功能设计

### 4.1 上传与画布初始化

- 上传后读取图片尺寸，按容器自适应显示。
- 初始化 Paper.js `Project` 与主 `Layer`。
- 底图放置在独立层并锁定，避免误选。

### 4.2 图形创建

- 矩形：`paper.Path.Rectangle`
- 圆/椭圆：`paper.Path.Ellipse`
- 三角形：通过 3 个顶点构建 `paper.Path`
- 五角星：按极坐标算法生成 10 个顶点（外半径/内半径交替）
- 心形：用贝塞尔曲线路径拼接闭合图形

### 4.3 元素选择与编辑

- 选中：点击命中测试（hitTest）
- 拖动：在 `select` 模式下拖拽更新位置
- 缩放：通过统一尺寸滑杆驱动 `item.scale`（保留比例）
- 删除：提供“删除选中”按钮
- 样式：
  - 填充色 `fillColor`
  - 描边色 `strokeColor`
  - 线宽 `strokeWidth`
  - 虚线 `dashArray`
  - 透明度 `opacity`

### 4.4 导出下载

- 导出方式：使用浏览器 `canvas.toDataURL('image/png')` 下载
- 文件命名：`annotated_<timestamp>.png`
- 导出前确保底图与标注处于同一渲染上下文

## 5. 数据结构与状态管理

- 页面内局部状态（先不进 Pinia）：
  - `activeTool`: 当前工具模式
  - `activeShapeType`: 当前绘制图形类型
  - `activeItemId`: 当前选中元素 id
  - `annotations`: 标注元数据列表（用于面板展示与后续扩展）
- Paper Item 自定义数据：
  - `data.id`
  - `data.kind`（rectangle/circle/triangle/star/heart）

## 6. UI 布局草案

- 顶部工具栏：返回、上传、工具切换、导出
- 左侧工具栏：图形快捷按钮（五种图形 + 选择工具）
- 中央画布区：Paper.js 画布
- 右侧属性面板：当前选中元素的样式/尺寸配置

## 7. 风险与对策

- 大图渲染性能：限制上传体积并对超大图提示压缩
- 命中精度问题：设置合理 hitTest 容差
- 心形/五角星缩放失真：统一以中心点等比缩放
- 移动端交互差异：首版优先桌面端，保留触控适配接口

## 8. 分步执行计划

1. 第一步：新增路由、任务页入口、标注页面骨架
2. 第二步：接入 Paper.js 与底图上传渲染
3. 第三步：实现图形创建（5 种）与选中拖拽
4. 第四步：实现属性编辑（颜色、线条、透明度、大小）
5. 第五步：实现导出下载与异常提示
6. 第六步：联调测试、补充文档、更新阶段进度

## 9. 验收标准（DoD）

- 用户可上传图片并在画布上创建 5 类图形标注
- 标注元素可选中、拖动、改色、改线宽/虚线、改透明度与尺寸
- 可导出包含标注结果的 PNG 文件，下载后内容与预览一致
- 页面在主流桌面浏览器可稳定运行，无明显控制台报错

## 10. 涉及文件（预期）

- `docs/image-annotation-implementation-plan.md`（本文件）
- `packages/frontend/src/router/index.ts`
- `packages/frontend/src/views/tasks/TasksView.vue`
- `packages/frontend/src/views/image-annotator/ImageAnnotatorView.vue`
- `packages/frontend/src/locales/zh.ts`
- `packages/frontend/src/locales/en.ts`
- `packages/frontend/package.json`（新增依赖）
