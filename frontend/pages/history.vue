<template>
  <div class="wrap">
    <h2 class="page">历史记录</h2>
    <p class="page-sub">Go 网关每 30 秒采集一次上游状态并落盘,此处展示本地历史曲线。</p>

    <div v-if="loading" class="loading">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}<br /><button class="retry" @click="load">重试</button></div>
    <template v-else>
      <div class="grid c3">
        <div class="card stat"><div class="v" style="color:var(--green)">{{ summary.uptime_pct }}%</div><div class="k">近 24h 可用率</div></div>
        <div class="card stat"><div class="v" style="color:var(--cyan)">{{ summary.avg_latency_ms }}<span class="dim"> ms</span></div><div class="k">平均延迟</div></div>
        <div class="card stat"><div class="v">{{ summary.total }}</div><div class="k">采集点数</div></div>
      </div>

      <div class="card" style="margin-top:16px">
        <h3>延迟曲线(ms,近 {{ points.length }} 个点)</h3>
        <svg :viewBox="`0 0 ${W} ${H}`" style="width:100%;height:220px">
          <polyline :points="line" fill="none" stroke="#22e6ff" stroke-width="2" />
          <circle v-for="(p, i) in dots" :key="i" :cx="p[0]" :cy="p[1]" r="3" :fill="p[2]" />
        </svg>
      </div>

      <div class="card" style="margin-top:16px">
        <h3>故障事件</h3>
        <div v-if="!incidents.length" class="dim">暂无异常事件,一切正常。</div>
        <table v-else class="tbl">
          <thead><tr><th>开始</th><th>持续</th><th>等级</th><th>说明</th><th>状态</th></tr></thead>
          <tbody>
            <tr v-for="i in incidents" :key="i.start_ts">
              <td>{{ new Date(i.start_ts).toLocaleString('zh-CN') }}</td>
              <td>{{ fmtDur(i.duration_sec) }}</td>
              <td>{{ i.worst_verdict }}</td>
              <td>{{ i.verdict_text }}</td>
              <td>{{ i.ongoing ? '进行中' : '已恢复' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="card" style="margin-top:16px">
        <h3>最近记录</h3>
        <table class="tbl">
          <thead><tr><th>时间</th><th>判定</th><th>延迟</th><th>服务正常</th><th>上报(异常/正常)</th></tr></thead>
          <tbody>
            <tr v-for="p in points.slice(0, 60)" :key="p.ts">
              <td>{{ timeAgo(p.ts) }}</td>
              <td>{{ p.verdict_text || p.verdict }}</td>
              <td>{{ p.latency_ms ?? '—' }} ms</td>
              <td>{{ p.ok_count }}/{{ p.ok_count + p.down_count }}</td>
              <td>{{ p.anomaly }}/{{ p.normal }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet, timeAgo } from '~/composables/useApi'

interface Point { ts: number; verdict: string; verdict_text: string; latency_ms: number | null; anomaly: number; normal: number; ok_count: number; down_count: number }

const W = 600, H = 200
const points = ref<Point[]>([])
const summary = ref<any>({})
const incidents = ref<any[]>([])
const loading = ref(true)
const error = ref('')

const line = computed(() => {
  const lat = points.value.map(p => p.latency_ms ?? 0)
  const max = Math.max(100, ...lat)
  return lat.map((v, i) => `${(i / Math.max(1, lat.length - 1)) * W},${H - 10 - (v / max) * (H - 30)}`).join(' ')
})
const dots = computed(() => {
  const lat = points.value.map(p => p.latency_ms ?? 0)
  const max = Math.max(100, ...lat)
  // echarts 太重,这里手绘:颜色按 verdict 区分
  return points.value.map((p, i) => [
    (i / Math.max(1, lat.length - 1)) * W,
    H - 10 - (lat[i] / max) * (H - 30),
    p.verdict === 'normal' ? '#2dffa8' : '#ff5470'
  ] as [number, number, string])
})

function fmtDur(sec: number): string {
  if (sec < 60) return `${sec} 秒`
  const m = Math.floor(sec / 60)
  if (m < 60) return `${m} 分钟`
  return `${Math.floor(m / 60)} 小时 ${m % 60} 分`
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [h, s, inc] = await Promise.all([
      apiGet<{ points: Point[] }>('/api/history', { limit: '288' }),
      apiGet<any>('/api/history/summary'),
      apiGet<{ incidents: any[] }>('/api/incidents').catch(() => ({ incidents: [] }))
    ])
    points.value = [...h.points].reverse() // 按时间正序画线
    summary.value = s
    incidents.value = inc.incidents
  } catch (e: any) {
    // 网关没开时给空态(历史依赖本地采集,无法兜底上游)
    error.value = '网关未连接:请先启动 Go 后端 (历史数据依赖本地采集)'
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>
