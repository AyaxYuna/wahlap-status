# 舞萌DX 状态监控(仿 isMaiDown)

Go 后端网关 + Nuxt 3 前端。上游文档:https://mai.chongxi.us/docs

数据来源于 isMaiDown by Chongxi(https://mai.chongxi.us),使用时必须标注来源。

## 目录

- `backend/` — Go 网关(标准库,零依赖):代理 + 缓存上游接口,定时采集落盘历史,透传上报。
- `frontend/` — Nuxt 3 前端:监控 / 历史 / 全服乐曲 / 全服排行 / 越级排行 / 关于。

## 快速启动

```powershell
# 终端 1:后端
cd backend
go build -o gateway.exe .
.\gateway.exe  # :8080,需联网访问上游

# 终端 2:前端
cd frontend
npm install
$env:NUXT_PUBLIC_API_BASE='http://localhost:8080'
npm run dev  # http://localhost:3000
```

环境变量见 `backend/.env.example` 与 `frontend/.env.example`。

## 说明

- 后端无第三方依赖,`go build` 即可编译;历史存在 `backend/data/history.jsonl`。
- 新增接口(详见 `backend/docs/API.md` 与 `GET /api/openapi.json`):
  `GET /api/overview`(聚合看板)、`GET /api/stream`(SSE 实时推送)、`GET /api/incidents`(故障事件)、
  `GET /api/latency/stats`(延迟分位)、`GET /api/reports/trend`(上报趋势)、`GET /api/history/range`(范围查询)。
- 前端直连上游兜底:网关未启动时,除历史页外其余页面仍可直连 `https://mai.chongxi.us` 显示数据。
- 本项目为非官方平台,与华立科技及世嘉官方无任何关联。
