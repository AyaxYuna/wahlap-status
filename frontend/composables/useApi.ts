// 统一 API 访问:先走 Go 网关,失败时直连上游兜底
export interface BotService {
  key: string
  name: string
  state: string
  state_text: string
  latency: number | null
  duration_text: string
}

export interface BotStatus {
  status: string
  status_text: string
  verdict: string
  verdict_text: string
  services: BotService[]
  summary: string
  latency: { current_ms: number; load_text: string; volatility_text: string }
  reports: { anomaly_count: number; normal_count: number }
  broadcast: { msg: string }
}

const UPSTREAM_PATHS: Record<string, string> = {
  '/api/status': '/api/bot',
  '/api/top100': '/api/top100',
  '/api/chart145': '/api/chart145',
  '/api/ratinglist': '/api/ratinglist'
}

export async function apiGet<T>(path: string, query?: Record<string, string>): Promise<T> {
  const { apiBase, upstreamBase } = useRuntimeConfig().public
  try {
    return await $fetch<T>(path, { baseURL: apiBase as string, query })
  } catch (e) {
    // 仅上游真实存在的接口才直连兜底;网关专有接口(历史/统计类)直接抛错,不打扰上游
    const upstream = UPSTREAM_PATHS[path]
    if (!upstream) throw e
    return await $fetch<T>(upstream, { baseURL: upstreamBase as string, query })
  }
}

/** 上游返回的封面是相对路径,拼成绝对地址 */
export function coverUrl(cover: string | null): string {
  if (!cover) return ''
  if (cover.startsWith('http')) return cover
  return 'https://mai.chongxi.us' + cover
}

export function fmtNum(n: number): string {
  return n.toLocaleString('zh-CN')
}

export function timeAgo(ts: number): string {
  const s = Math.max(1, Math.floor((Date.now() - ts) / 1000))
  if (s < 60) return `${s} 秒前`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m} 分钟前`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h} 小时前`
  return `${Math.floor(h / 24)} 天前`
}
