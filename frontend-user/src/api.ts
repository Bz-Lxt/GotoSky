const BASE = ''

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const r = await fetch(BASE + path, { headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) }, ...init })
  const j = await r.json().catch(() => ({}))
  if (!r.ok) throw new Error(j?.error?.message || r.statusText)
  return (j.data ?? j) as T
}

export const api = {
  health: () => req<{ status: string }>('/health'),
  meta: () => req<any>('/api/v1/meta'),
  quota: () => req<any>('/api/v1/quota'),
  sites: () => req<any[]>('/api/v1/sites'),
  createSite: (b: any) => req('/api/v1/sites', { method: 'POST', body: JSON.stringify(b) }),
  targets: (q = '') => req<any[]>(`/api/v1/targets?q=${encodeURIComponent(q)}`),
  scores: (site: string, target = '') => req<any[]>(`/api/v1/scores?site_id=${site}&target_id=${target}`),
  windows: (site: string) => req<any[]>(`/api/v1/windows?site_id=${site}`),
  heatmap: (site: string, target = '') => req<any[]>(`/api/v1/heatmap?site_id=${site}&target_id=${target}`),
  refresh: (site: string) => req('/api/v1/forecast/refresh', { method: 'POST', body: JSON.stringify({ site_id: site }) }),
  skytrack: (site: string, target: string) => req<any[]>(`/api/v1/skytrack?site_id=${site}&target_id=${target}`),
  night: (site: string) => req<any>(`/api/v1/night?site_id=${site}`),
  mystery: (site: string) => req<any[]>(`/api/v1/mystery?site_id=${site}`),
  plans: (site: string) => req<any[]>(`/api/v1/plans?site_id=${site}`),
  createPlan: (b: any) => req('/api/v1/plans', { method: 'POST', body: JSON.stringify(b) }),
  addItem: (id: string, b: any) => req(`/api/v1/plans/${id}/items`, { method: 'POST', body: JSON.stringify(b) }),
  equipment: () => req<any[]>('/api/v1/equipment'),
  rigs: () => req<any[]>('/api/v1/rigs'),
  sessions: () => req<any[]>('/api/v1/sessions'),
  startSession: (b: any) => req('/api/v1/sessions', { method: 'POST', body: JSON.stringify(b) }),
  cmd: (id: string, verb: string, payload: any = {}) => req(`/api/v1/sessions/${id}/commands`, { method: 'POST', body: JSON.stringify({ verb, payload, command_id: crypto.randomUUID() }) }),
  stars: () => req<any[]>('/api/v1/stars'),
  login: (username: string, password: string) => req('/api/v1/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  alerts: () => req<any[]>('/api/v1/alerts'),
}

export function fmt(ts: string) {
  const d = new Date(ts)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

export function tierLabel(t: string) {
  return ({ GOLD: '金', SILVER: '银', BRONZE: '铜', POOR: '平', UNUSABLE: '不可拍' } as any)[t] || t
}
