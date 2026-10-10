<template>
  <div class="wrap">
    <div v-if="broadcast" class="notice">📢 {{ broadcast }}</div>

    <div v-if="loading" class="loading">正在连接探针…</div>
    <div v-else-if="error" class="error">
      获取状态失败:{{ error }}
      <br /><button class="retry" @click="load">重试</button>
    </div>
    <template v-else-if="st">
      <div class="hero" :class="isOk ? 'ok' : 'bad'">
        <h1>{{ st.verdict_text }}</h1>
        <p class="sub">{{ st.summary }}</p>
        <p class="dim" style="margin-top:14px">
          {{ updatedText }} · {{ countdown }}s 后自动刷新
          <a href="#" @click.prevent="load" style="color:var(--cyan)">立即刷新</a>
        </p>
      </div>

      <div class="grid c6" style="margin-top:16px">
        <div v-for="s in st.services" :key="s.key" class="card svc" :class="s.state">
          <div class="name">{{ s.name }}</div>
          <div class="lat">{{ s.latency ?? '—' }} <small>ms</small></div>
          <span class="pill" :class="s.state">{{ s.state_text }}</span>
          <div v-if="s.duration_text" class="dim" style="margin-top:6px">{{ s.duration_text }}</div>
        </div>
      </div>

      <div class="grid c3" style="margin-top:16px">
        <div class="card stat">
          <div class="v" style="color:var(--cyan)">{{ st.latency.current_ms }}<span class="dim"> ms</span></div>
          <div class="k">当前延迟 · 负载{{ st.latency.load_text }} · {{ st.latency.volatility_text }}</div>
        </div>
        <div class="card stat">
          <div class="v" style="color:var(--red)">{{ st.reports.anomaly_count }}</div>
          <div class="k">近 1 小时异常上报</div>
        </div>
        <div class="card stat">
          <div class="v" style="color:var(--green)">{{ st.reports.normal_count }}</div>
          <div class="k">近 1 小时正常上报</div>
        </div>
      </div>

      <div v-if="avail" class="card" style="margin-top:16px">
        <h3>服务可用性 · 近 {{ avail.hours }} 小时</h3>
        <svg :viewBox="`0 0 ${BW} ${BH}`" style="width:100%;height:130px">
          <rect v-for="(b, i) in avail.buckets" :key="i"
            :x="i * (BW / avail.buckets.length) + 1" :y="BH - 18 - b.ratio * (BH - 28)"
            :width="BW / avail.buckets.length - 2" :height="Math.max(2, b.ratio * (BH - 28))" rx="2"
            :fill="b.ratio >= 0.995 ? '#2dffa8' : b.ratio >= 0.9 ? '#ffb020' : '#ff5470'" />
          <text x="0" :y="BH - 4" fill="#9aa3c7" font-size="10">-{{ avail.hours }}h</text>
          <text :x="BW / 2 - 12" :y="BH - 4" fill="#9aa3c7" font-size="10">-{{ avail.hours / 2 }}h</text>
          <text :x="BW - 30" :y="BH - 4" fill="#9aa3c7" font-size="10">NOW</text>
        </svg>
        <div class="grid c2" style="margin-top:8px">
          <div v-for="s in avail.services" :key="s.key" class="avail-row">
            <span>{{ s.name }}</span>
            <b :style="{ color: pctColor(s.availability_pct) }">{{ s.availability_pct }}%</b>
            <div class="dotbar" :style="{ width: s.availability_pct + '%', color: pctColor(s.availability_pct) }"></div>
          </div>
        </div>
        <div v-if="!avail.services.length" class="dim">暂无数据,网关采集一段时间后展示。</div>
      </div>

      <div v-if="hasChart" class="card" style="margin-top:16px">
        <h3>图表概览</h3>
        <div class="tabs" style="margin-bottom:10px">
          <button :class="{ on: chartMode === 'lat' }" @click="chartMode = 'lat'">延迟</button>
          <button :class="{ on: chartMode === 'rep' }" @click="chartMode = 'rep'">上报</button>
        </div>
        <div class="legend">
          <span v-for="(l, i) in chartLegend" :key="l.key"><i :style="{ background: l.color }"></i>{{ l.name }}</span>
        </div>
        <svg :viewBox="`0 0 ${CW} ${CH}`" style="width:100%;height:260px">
          <polyline v-for="(p, i) in chartLines" :key="i" :points="p.d" fill="none" :stroke="p.color" stroke-width="2" />
          <text :x="CW - 8" y="14" fill="#9aa3c7" font-size="11" text-anchor="end">{{ chartMaxLabel }}</text>
          <text x="0" :y="CH - 4" fill="#9aa3c7" font-size="10">{{ chartT0 }}</text>
          <text :x="CW - 30" :y="CH - 4" fill="#9aa3c7" font-size="10">NOW</text>
        </svg>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet, type BotStatus } from '~/composables/useApi'

const st = ref<BotStatus | null>(null)
const loading = ref(true)
const error = ref('')
const updatedAt = ref(0)
const countdown = ref(30)
let timer: ReturnType<typeof setInterval> | null = null

const isOk = computed(() => st.value?.verdict === 'normal')
const broadcast = computed(() => st.value?.broadcast?.msg || '')
const updatedText = computed(() =>
  updatedAt.value ? `更新于 ${new Date(updatedAt.value).toLocaleTimeString('zh-CN')}` : ''
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    st.value = await apiGet<BotStatus>('/api/status')
    updatedAt.value = Date.now()
    countdown.value = 30
  } catch (e: any) {
    error.value = e?.message || '网络错误'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  loadExtra()
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) load()
  }, 1000)
})
onUnmounted(() => { if (timer) clearInterval(timer) })

/* ---- 服务可用性 + 图表概览(依赖网关本地采集,失败则隐藏) ---- */
const BW = 600, BH = 130, CW = 900, CH = 250
const avail = ref<{ hours: number; buckets: { ts: number; ratio: number }[]; services: any[] } | null>(null)
const latSeries = ref<{ key: string; name: string; series: [number, number][] }[]>([])
const repTrend = ref<{ ts: number; anomaly: number; normal: number }[]>([])
const chartMode = ref<'lat' | 'rep'>('lat')
const SVC_COLORS = ['#22e6ff', '#ffd84d', '#ff9d3e', '#ff3ec8', '#9d7bff', '#2dffa8', '#7affd4', '#ff5470']
const svcColor = (i: number) => SVC_COLORS[i % SVC_COLORS.length]

function pctColor(p: string | number): string {
  const v = Number(p)
  return v >= 99.5 ? '#2dffa8' : v >= 90 ? '#ffb020' : '#ff5470'
}
function hhmm(ts: number): string {
  return new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function loadExtra() {
  try {
    const [a, l, t] = await Promise.all([
      apiGet<any>('/api/services/availability', { hours: '12' }),
      apiGet<any>('/api/services/latency', { hours: '24' }),
      apiGet<any>('/api/reports/trend', { limit: '240' })
    ])
    avail.value = a
    latSeries.value = l.services || []
    repTrend.value = t.trend || []
  } catch {
    // 网关未启动:仅隐藏新增区块,基础状态仍可走上游兜底
  }
}

const hasChart = computed(() => latSeries.value.length > 0 || repTrend.value.length > 0)
const chartLegend = computed(() => {
  if (chartMode.value === 'lat') return latSeries.value.map((s, i) => ({ key: s.key, name: s.name, color: svcColor(i) }))
  return [{ key: 'anomaly', name: '异常上报', color: '#ff5470' }, { key: 'normal', name: '正常上报', color: '#2dffa8' }]
})
const chartData = computed(() => {
  if (chartMode.value === 'lat') {
    return latSeries.value.map((s, i) => ({ color: svcColor(i), pts: s.series }))
  }
  return [
    { color: '#ff5470', pts: repTrend.value.map(p => [p.ts, p.anomaly] as [number, number] ) },
    { color: '#2dffa8', pts: repTrend.value.map(p => [p.ts, p.normal] as [number, number] ) }
  ]
})
const chartLines = computed(() => {
  const all = chartData.value.flatMap(d => d.pts)
  if (!all.length) return []
  const ts = all.map(p => p[0])
  const t0 = Math.min(...ts)
  const t1 = Math.max(Date.now(), ...ts)
  const maxV = Math.max(100, ...all.map(p => p[1]))
  return chartData.value.map(d => ({
    color: d.color,
    d: d.pts.map(([t, v]) => `${((t - t0) / Math.max(1, t1 - t0)) * CW},${CH - 20 - (v / maxV) * (CH - 50)}`).join(' ')
  }))
})
const chartMaxLabel = computed(() => {
  const all = chartData.value.flatMap(d => d.pts)
  if (!all.length) return ''
  const maxV = Math.max(100, ...all.map(p => p[1]))
  return chartMode.value === 'lat' ? `${maxV}ms` : `${maxV} 条`
})
const chartT0 = computed(() => {
  const all = chartData.value.flatMap(d => d.pts)
  return all.length ? hhmm(Math.min(...all.map(p => p[0]))) : ''
})
</script>
