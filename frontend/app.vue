<template>
  <div>
    <nav class="nav">
      <div class="nav-inner">
        <NuxtLink to="/" class="brand">
          <span class="brand-dot" :class="dotClass"></span>
          华立服务器死了吗?
        </NuxtLink>
        <div class="links">
          <NuxtLink to="/">监控</NuxtLink>
          <NuxtLink to="/history">历史</NuxtLink>
          <NuxtLink to="/top100">全服乐曲</NuxtLink>
          <NuxtLink to="/ratinglist">全服排行</NuxtLink>
          <NuxtLink to="/chart145">越级排行</NuxtLink>
          <NuxtLink to="/about">关于/文档</NuxtLink>
        </div>
      </div>
    </nav>
    <NuxtPage />
    <footer class="footer">
      数据来源于 <a href="https://mai.chongxi.us" target="_blank" rel="noopener">isMaiDown by Chongxi</a>
      · 本站为非官方平台,与华立科技及世嘉官方无任何关联
    </footer>
  </div>
</template>

<script setup lang="ts">
// 顶栏圆点跟随实时状态变色
const { data } = await useFetch('/api/status', {
  baseURL: useRuntimeConfig().public.apiBase,
  server: false,
  lazy: true,
  retry: 0
})
const dotClass = computed(() => {
  const v = (data.value as any)?.verdict
  if (v === 'normal' || v === undefined) return ''
  if (v === 'outage') return 'down'
  return 'warn'
})
</script>
