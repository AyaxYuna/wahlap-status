<template>
  <div class="wrap">
    <div class="loading" v-if="!error">正在完成落雪授权…</div>
    <div class="error" v-else>{{ error }}<br /><NuxtLink to="/player" style="color:var(--cyan)">返回重试</NuxtLink></div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string
  if (!code || !state) {
    error.value = '缺少授权码,请重新发起登录。'
    return
  }
  try {
    const d = await $fetch<any>('/api/lxns/callback', {
      baseURL: useRuntimeConfig().public.apiBase as string,
      query: { code, state }
    })
    localStorage.setItem('lxns_session', d.session)
    await navigateTo('/player')
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || '授权失败'
  }
})
</script>
