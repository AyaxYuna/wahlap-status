<template>
  <div class="wrap">
    <h2 class="page">越级排行 · 定数 14.5–15.0</h2>
    <p class="page-sub">合计游玩 <b style="color:var(--gold)">{{ fmtNum(total) }}</b> 人次 · 按游玩人数排序</p>

    <div v-if="loading" class="loading">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}<br /><button class="retry" @click="load">重试</button></div>
    <template v-else>
      <div class="toolbar">
        <input v-model="q" class="search" placeholder="搜索曲名…" />
        <select v-model="level" class="sel">
          <option value="">全部难度</option>
          <option value="4">MASTER</option>
          <option value="3">EXPERT</option>
          <option value="0">宴/狂</option>
        </select>
      </div>
      <div class="card">
        <div v-for="s in filtered" :key="s.musicId + '-' + s.level" class="song">
          <div class="rank" :class="s.rank <= 3 ? 'r' + s.rank : ''">{{ s.rank }}</div>
          <img v-if="s.cover" class="cover" :src="coverUrl(s.cover)" :alt="s.name" loading="lazy" />
          <div>
            <div class="t">{{ s.name }}<span class="badge" :class="badgeCls(s.level)">{{ levelText(s.level) }}</span></div>
            <div class="s">定数 {{ s.rate }} · 平均 {{ s.avg }}% · 中位 {{ s.med }}% · 众数 {{ s.mode }}%</div>
          </div>
          <div class="plays">{{ fmtNum(s.players) }}<small>人游玩</small></div>
        </div>
        <div v-if="!filtered.length" class="loading">没有匹配的谱面</div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet, coverUrl, fmtNum } from '~/composables/useApi'

interface Chart { rank: number; players: number; musicId: number; name: string; level: number; rate: number; avg: number; med: number; mode: number; cover: string | null }
const songs = ref<Chart[]>([])
const total = ref(0)
const q = ref('')
const level = ref('')
const loading = ref(true)
const error = ref('')

const filtered = computed(() => songs.value.filter(s =>
  (!q.value.trim() || s.name.toLowerCase().includes(q.value.trim().toLowerCase())) &&
  (level.value === '' || s.level === Number(level.value))
))
function levelText(l: number) { return l === 4 ? 'MAS' : l === 3 ? 'EXP' : '宴/狂' }
function badgeCls(l: number) { return l === 4 ? 'mas' : l === 3 ? 'exp' : 'uto' }

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await apiGet<{ totalPlayers: number; songs: Chart[] }>('/api/chart145')
    total.value = d.totalPlayers
    songs.value = d.songs
  } catch (e: any) {
    error.value = e?.message || '网络错误'
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>
