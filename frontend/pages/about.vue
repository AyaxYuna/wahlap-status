<template>
  <div class="wrap">
    <h2 class="page">关于 / 接口文档</h2>
    <p class="page-sub">本站为非官方平台,数据不代表华立科技官方立场,与华立科技及世嘉官方无任何关联。</p>

    <div class="grid c2">
      <div class="card">
        <h3>数据来源</h3>
        <p>平台数据由机台采集、服务器探针、玩家上报、Bot 关键词监听四路实时聚合,源自
          <a href="https://mai.chongxi.us" target="_blank" rel="noopener" style="color:var(--cyan)">isMaiDown by Chongxi</a>。
        </p>
        <p class="dim">本站数据免费公开,可自由使用 / 分发 / 商业使用,需标注「数据来源于 isMaiDown by Chongxi」。</p>
        <p class="dim">QQ 群:176358258 · 邮箱:qwq@chongxi.us</p>
      </div>
      <div class="card">
        <h3>本项目架构</h3>
        <p><code class="inline">backend/</code> Go 网关:代理 + 缓存上游接口,每 30 秒采集状态落盘为本地历史,并透传上报签名。</p>
        <p><code class="inline">frontend/</code> Nuxt 3 前端:本页面即由 Nuxt 渲染,API 基址通过 <code class="inline">NUXT_PUBLIC_API_BASE</code> 配置。</p>
        <p class="dim">网关不可用时,前端会自动直连 <code class="inline">https://mai.chongxi.us</code> 兜底(历史页除外)。</p>
      </div>
    </div>

    <div class="card" style="margin-top:16px">
      <h3>上游公开接口(https://mai.chongxi.us)</h3>
      <table class="tbl">
        <thead><tr><th>方法</th><th>路径</th><th>说明</th></tr></thead>
        <tbody>
          <tr><td>GET</td><td><code class="inline">/api/bot</code></td><td>服务器状态与延迟,无需认证</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/top100</code></td><td>全服游玩次数 Top100</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/ratinglist[?version=dx2026]</code></td><td>Rating 百分一段位分布</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/chart145</code></td><td>定数 14.5–15.0 谱面排行</td></tr>
          <tr><td>POST</td><td><code class="inline">https://maiapi.chongxi.us/bot-post</code></td><td>机台异常上报,需 HMAC-SHA256 签名,建议 30s 聚合一次</td></tr>
        </tbody>
      </table>
    </div>

    <div class="card" style="margin-top:16px">
      <h3>本网关接口(默认 :8080)</h3>
      <table class="tbl">
        <thead><tr><th>方法</th><th>路径</th><th>说明</th></tr></thead>
        <tbody>
          <tr><td>GET</td><td><code class="inline">/healthz</code></td><td>健康检查</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/status</code></td><td>状态(兼容 <code class="inline">/api/bot</code>),20s 缓存</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/top100 · /api/chart145 · /api/ratinglist</code></td><td>静态快照,1h 缓存</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/history?limit=288</code></td><td>本地采集历史(最新在前)</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/history/summary</code></td><td>近 24h 可用率 / 平均延迟</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/history/range?from=&amp;to=</code></td><td>按毫秒时间戳范围查历史</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/incidents</code></td><td>故障事件列表(异常连续段自动提取)</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/latency/stats?hours=24</code></td><td>延迟分位:min/p50/p90/p99/max/avg</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/reports/trend</code></td><td>异常/正常上报趋势</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/overview</code></td><td>聚合看板:状态 + 摘要 + 分位 + 进行中事件</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/stream</code></td><td>SSE 实时推送(EventSource 订阅)</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/openapi.json</code></td><td>OpenAPI 3.0 规格</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/lxns/auth-url</code></td><td>落雪 OAuth 授权链接</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/lxns/me · /api/lxns/scores</code></td><td>登录用户与成绩(需 X-Session)</td></tr>
          <tr><td>GET</td><td><code class="inline">/api/lxns/b50</code></td><td>Best50(旧35+新15,需 X-Session)</td></tr>
          <tr><td>POST</td><td><code class="inline">/api/lxns/scores</code></td><td>成绩上传(需 write_player 权限)</td></tr>
          <tr><td>POST</td><td><code class="inline">/api/report</code></td><td>上报透传(需带 X-Client-ID / X-Timestamp / X-Signature)</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
