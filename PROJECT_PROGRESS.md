# R-Comic 项目进展追踪

> 本文档用于跟踪 R-Comic（Wails v2: Go + Vue3）项目的目标、状态、里程碑、待办与关键决策，方便多人协作与持续推进。

## 1. 项目目标（当前理解）
- 基于 Wails 的桌面应用：管理与浏览来自文件夹或 ZIP 压缩包的漫画。
- 支持拖放（文件夹 / ZIP）导入；自动提取首图；展示列表；搜索与删除。
- 兼容中文/多语言文件名，自然排序，ZIP 内图片直读与 HTTP 流式输出。

## 2. 当前状态概览
- 后端（Go）
  - `main.go`：Wails 启动与自定义 HTTP `FileLoader`，支持 `zipPath!imagePath` 直读与流式输出。
  - `app.go`：数据库初始化（SQLite：`comics`/`images`），拖放处理、ZIP/文件夹广度优先找图、自然排序、Base64 编码输出、CRUD 与搜索。
  - `comic.db`：本地 SQLite 文件已存在。
- 前端（Vue3 + Vite）
  - `frontend/src/App.vue`：加载/展示漫画列表、拖放初始化、搜索、删除。
  - `frontend/src/components/*`：`ComicCard`、`SearchToolbar`、`HelloWorld` 可用。
  - `frontend/src/dragAndDrop.js`：集成 Wails 的 `OnFileDrop`。
  - Wails 绑定位于 `frontend/wailsjs/`（自动生成）。
- 构建
  - Wails 与 Vite 已配置，`wails.json` 与 `vite.config.js` 正常。

## 3. 已实现功能
- 拖放文件夹/ZIP；ZIP/文件夹下图片的 BFS 搜索与自然排序。
- 数据库存储与查询：新增/覆盖、获取列表、关键字搜索、删除。
- 图片服务
  - 方式A：后端 HTTP `FileLoader` 读本地文件或 ZIP 流式输出（`zip!path`）。
  - 方式B：`GetImageBase64` 返回 Base64 DataURL。
- 中文路径处理：对 ZIP 内部路径做 URL 解码尝试与模糊匹配。

## 4. 近期里程碑
- M1：稳定的导入与列表显示
  - 拖放多文件稳定；导入进度与错误提示。
  - 图片预览稳定显示（Base64 或直链二选一统一策略）。
- M2：阅读器基础（单本预览）
  - 点击卡片进入阅读器，支持上一张/下一张、缩放与旋转、键盘快捷键。
- M3：库管理增强
  - 批量删除、批量导入、重复项检测（基于路径/哈希）。
- M4：体验与发布
  - 应用设置（缓存、缩略图策略、语言）；打包发布；崩溃日志与错误上报。

## 5. 待办清单（Backlog）
- 前端
  - 统一图片加载方案：优先使用 `comic://zip!path` 直链或 Base64，减少重复逻辑。
  - 在 `ComicCard` 显示首图缩略图（当前仅有数据层支持）。
  - 导入/搜索的加载态与错误提示更友好；空态引导按钮（选择文件夹）。
  - 引入路由与阅读页（`/reader/:id`）。
- 后端
  - `GetImageData` 支持 ZIP 路径（当前要求本地文件路径，和 `FileLoader` 行为不一致）。
  - 首图提取时生成/缓存缩略图以提速列表渲染。
  - `images` 表写入每页图片记录（当前未落表）。
  - 重复检测与并发导入安全。
- 基建
  - 日志与调试级别控制；单元测试（自然排序、BFS）；CI 构建；README 更新。

## 6. 风险与技术债
- 双通道图片服务（HTTP 与 Base64）并存，易造成前后端耦合与分歧。
- ZIP 中文名/编码兼容性：当前用 URL 解码与模糊匹配，边界场景仍可能失败。
- 数据一致性：`images` 表暂未填充，后续特性（阅读器/分页）需要完善。

## 7. 关键决策记录（变更请更新此表）
- 采用 `zipPath!imagePath` 地址格式，由自定义 `FileLoader` 直出图片；同时保留 Base64 渲染备用。
- BFS + 自然排序，保证“第1张”更符合人类直觉。

## 8. 发布与运行检查清单
- 开发
  - `wails dev` 可启动；前端热更新可用；拖放导入成功；列表可见。
- 构建
  - `wails build` 产物可运行；缩略图加载速度达标；基础功能回归通过。
- 文档
  - 更新 `README` 使用说明；`PROJECT_PROGRESS.md` 反映最新状态。

---
维护人：@Bailin
最后更新：2025-08-08
