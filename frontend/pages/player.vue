<template>
  <div class="wrap">
    <h2 class="page">我的成绩</h2>
    <p class="page-sub">通过落雪查分器 OAuth 登录后查看玩家信息与成绩,密钥只保存在 Go 网关侧。</p>

    <div v-if="loading" class="loading">加载中…</div>

    <div v-else-if="!session" class="card" style="text-align:center;padding:48px 20px">
      <div style="font-size:44px">🎧</div>
      <h3 style="font-size:18px;color:var(--text)">绑定落雪查分器账号</h3>
      <p class="dim">将跳转到 maimai.lxns.net 授权,本站不会看到你的查分器密码。</p>
      <button class="retry" @click="login">使用落雪查分器登录</button>
      <div v-if="error" class="error">{{ error }}</div>
    </div>

    <template v-else>
      <div v-if="error" class="error">{{ error }}<br /><button class="retry" @click="load">重试</button></div>
      <div v-if="player" class="card" style="display:flex;gap:18px;align-items:center">
        <img v-if="player.icon?.id" class="cover" style="width:72px;height:72px"
          :src="`https://assets2.lxns.net/maimai/icon/${player.icon.id}.png`" alt="icon" loading="lazy" />
        <div>
          <div style="font-size:22px;font-weight:900">{{ player.name }}</div>
          <div class="dim">好友码 {{ player.friend_code }} · 觉醒 {{ player.star }}</div>
        </div>
        <div class="plays" style="margin-left:auto;text-align:right;font-weight:800;color:var(--gold);font-size:28px">
          {{ player.rating }}<small style="display:block;color:var(--muted);font-weight:400;font-size:12px">DX Rating</small>
        </div>
        <button class="retry" style="background:none;border:1px solid var(--card-border);color:var(--muted)" @click="logout">退出</button>
      </div>

      <div v-if="scores.length" class="card" style="margin-top:16px">
        <h3>Best 50(按 DX Rating 排序) · 共 {{ scores.length }} 条成绩</h3>
        <div class="toolbar"><input v-model="q" class="search" placeholder="搜索曲名…" /></div>
        <div v-for="s in filtered.slice(0, 50)" :key="s.id + '-' + s.type + '-' + s.level_index" class="song">
          <img class="cover" :src="`https://assets2.lxns.net/maimai/jacket/${s.id}.png`" :alt="s.song_name" loading="lazy" />
          <div>
            <div class="t">{{ s.song_name }}<span class="badge" :class="typeCls(s.type)">{{ typeText(s) }}</span></div>
            <div class="s">{{ s.level }} · 达成率 {{ s.achievements }}% · DX {{ s.dx_score }}{{ s.fc ? ' · ' + String(s.fc).toUpperCase() : '' }}{{ s.fs ? ' · ' + String(s.fs).toUpperCase() : '' }}</div>
          </div>
          <div class="plays">{{ Math.floor(s.dx_rating) }}<small>{{ String(s.rate || '').toUpperCase() }}</small></div>
        </div>
      </div>
      <div v-else-if="player" class="card" style="margin-top:16px"><div class="dim">暂无成绩数据。</div></div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet } from '~/composables/useApi'

const LEVEL_NAMES = ['BASIC', 'ADV', 'EXP', 'MAS', 'ReMAS']
const session = ref<string>(process.client ? localStorage.getItem('lxns_session') || '' : '')
const player = ref<any>(null)
const scores = ref<any[]>([])
const q = ref('')
const loading = ref(true)
const error = ref('')

const filtered = computed(() => {
  const k = q.value.trim().toLowerCase()
  const sorted = [...scores.value].sort((a, b) => (b.dx_rating || 0) - (a.dx_rating || 0))
  if (!k) return sorted
  return sorted.filter(s => String(s.song_name || '').toLowerCase().includes(k))
})
function typeText(s: any) {
  const lv = LEVEL_NAMES[s.level_index] || s.level_index
  return `${s.type === 'dx' ? 'DX' : s.type === 'utage' ? '宴' : 'STD'} ${lv}`
}
function typeCls(t: string) { return t === 'dx' ? 'mas' : t === 'utage' ? 'uto' : 'exp' }

async function login() {
  error.value = ''
  try {
    const d = await apiGet<{ auth_url: string }>('/api/lxns/auth-url')
    location.href = d.auth_url
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || '网关未配置落雪应用,联系站长填写 LXNS_CLIENT_ID'
  }
}
async function load() {
  if (!session.value) { loading.value = false; return }
  loading.value = true
  error.value = ''
  try {
    const me = await $fetch<any>('/api/lxns/me', {
      baseURL: useRuntimeConfig().public.apiBase as string,
      headers: { 'X-Session': session.value }
    })
    player.value = me.player
    const sc = await $fetch<any>('/api/lxns/scores', {
      baseURL: useRuntimeConfig().public.apiBase as string,
      headers: { 'X-Session': session.value }
    })
    scores.value = sc?.data || []
  } catch (e: any) {
    if (e?.response?.status === 401) {
      session.value = ''
      if (process.client) localStorage.removeItem('lxns_session')
      error.value = ''
    } else {
      error.value = e?.data?.error || e?.message || '加载失败'
    }
  } finally {
    loading.value = false
  }
}
async function logout() {
  try {
    await $fetch('/api/lxns/logout', {
      method: 'POST', baseURL: useRuntimeConfig().public.apiBase as string,
      headers: { 'X-Session': session.value }
    })
  } catch {}
  session.value = ''
  player.value = null
  scores.value = []
  if (process.client) localStorage.removeItem('lxns_session')
}
onMounted(load)
</script>
