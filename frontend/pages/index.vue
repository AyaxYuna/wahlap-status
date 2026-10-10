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
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) load()
  }, 1000)
})
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>
