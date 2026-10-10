<template>
  <div class="wrap">
    <h2 class="page">全服乐曲 Top100</h2>
    <p class="page-sub">按全服游玩次数排序 · 全服总游玩 <b style="color:var(--gold)">{{ fmtNum(total) }}</b> 次</p>

    <div v-if="loading" class="loading">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}<br /><button class="retry" @click="load">重试</button></div>
    <template v-else>
      <div class="toolbar"><input v-model="q" class="search" placeholder="搜索曲名…" /></div>
      <div class="card">
        <div v-for="s in filtered" :key="s.id" class="song">
          <div class="rank" :class="s.rank <= 3 ? 'r' + s.rank : ''">{{ s.rank }}</div>
          <img class="cover" :src="coverUrl(s.cover)" :alt="s.name" loading="lazy" />
          <div><div class="t">{{ s.name }}</div><div class="s">ID {{ s.id }}</div></div>
          <div class="plays">{{ fmtNum(s.plays) }}<small>占 {{ pct(s.plays) }}%</small></div>
        </div>
        <div v-if="!filtered.length" class="loading">没有匹配的乐曲</div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet, coverUrl, fmtNum } from '~/composables/useApi'

interface Song { id: number; name: string; plays: number; cover: string }
const songs = ref<(Song & { rank: number })[]>([])
const total = ref(0)
const q = ref('')
const loading = ref(true)
const error = ref('')

const filtered = computed(() => {
  const k = q.value.trim().toLowerCase()
  if (!k) return songs.value
  return songs.value.filter(s => s.name.toLowerCase().includes(k) || String(s.id).includes(k))
})
function pct(n: number) {
  return total.value ? ((n / total.value) * 100).toFixed(2) : '0.00'
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await apiGet<{ totalPlays: number; songs: Song[] }>('/api/top100')
    total.value = d.totalPlays
    songs.value = d.songs.map((s, i) => ({ ...s, rank: i + 1 }))
  } catch (e: any) {
    error.value = e?.message || '网络错误'
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>
