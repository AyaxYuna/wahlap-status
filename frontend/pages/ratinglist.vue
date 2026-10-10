<template>
  <div class="wrap">
    <h2 class="page">全服 Rating 分布</h2>
    <p class="page-sub">每 100 分一段 · 当前版本总玩家 <b style="color:var(--gold)">{{ fmtNum(cur.total) }}</b></p>

    <div v-if="loading" class="loading">加载中…</div>
    <div v-else-if="error" class="error">{{ error }}<br /><button class="retry" @click="load">重试</button></div>
    <template v-else>
      <div class="tabs" style="margin-bottom:16px">
        <button v-for="v in versions" :key="v.key" :class="{ on: v.key === curKey }" @click="curKey = v.key">
          {{ v.label }} ({{ fmtNum(v.total) }})
        </button>
      </div>
      <div class="card">
        <h3>{{ cur.label }} · 每 100 Rating 一段</h3>
        <div v-for="b in cur.data" :key="b.bucket" class="bar-row">
          <span class="dim">{{ b.bucket }}–{{ b.bucket + 99 }}</span>
          <div class="bar-track"><div class="bar-fill" :style="{ width: (b.count / maxCount * 100).toFixed(1) + '%' }"></div></div>
          <span style="text-align:right">{{ fmtNum(b.count) }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { apiGet, fmtNum } from '~/composables/useApi'

interface Bucket { bucket: number; count: number }
interface Ver { key: string; label: string; total: number; data: Bucket[] }
const versions = ref<Ver[]>([])
const curKey = ref('')
const loading = ref(true)
const error = ref('')

const cur = computed<Ver>(() => versions.value.find(v => v.key === curKey.value) || { key: '', label: '', total: 0, data: [] })
const maxCount = computed(() => Math.max(1, ...cur.value.data.map(b => b.count)))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await apiGet<{ versions: any[]; data: Ver[] }>('/api/ratinglist')
    versions.value = d.data
    curKey.value = d.data[d.data.length - 1]?.key || ''
  } catch (e: any) {
    error.value = e?.message || '网络错误'
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>
