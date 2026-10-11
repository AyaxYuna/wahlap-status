// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-11-01',
  devtools: { enabled: false },
  modules: ['@nuxt/ui'],
  devServer: { port: 9561 },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      title: '华立服务器死了吗？— 舞萌DX 状态监控',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '舞萌DX 服务器实时监测：探针状态、延迟、全服排行。数据来源于 isMaiDown by Chongxi。' }
      ]
    }
  },
  runtimeConfig: {
    public: {
      // Go 网关地址,生产环境通过 NUXT_PUBLIC_API_BASE 覆盖
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9560',
      // 网关不可用时直连上游兜底
      upstreamBase: 'https://mai.chongxi.us',
      // 手动上报签名密钥(发邮件至 qwq@chongxi.us 获取后填入环境变量)
      reportClientId: process.env.NUXT_PUBLIC_REPORT_CLIENT_ID || '<your-key>',
      reportPrivateKey: process.env.NUXT_PUBLIC_REPORT_PRIVATE_KEY || '<your-key>'
    }
  },
  colorMode: {
    preference: 'system',
    fallback: 'dark',
    classSuffix: ''
  }
})
