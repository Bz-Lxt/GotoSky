import { NavLink, Route, Routes } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from './api'
import { ToastHost, pushToast } from './ui/Toast'
import { CalendarPage } from './pages/Calendar'
import { HeatmapPage } from './pages/Heatmap'
import { SkyPage } from './pages/Sky'
import { WallPage } from './pages/Wall'
import { MysteryPage } from './pages/Mystery'
import { GearPage } from './pages/Gear'
import { PlansPage } from './pages/Plans'

export function App() {
  const [sites, setSites] = useState<any[]>([])
  const [site, setSite] = useState('')
  const [quota, setQuota] = useState<any>(null)
  const [meta, setMeta] = useState<any>(null)

  useEffect(() => {
    api.sites().then((xs) => { setSites(xs); if (xs[0]) setSite(xs[0].id) }).catch((e) => pushToast(e.message))
    api.quota().then(setQuota).catch(() => {})
    api.meta().then(setMeta).catch(() => {})
  }, [])

  const nav = [
    ['/', '夜间日历'],
    ['/heatmap', '天气矩阵'],
    ['/sky', '对焦星图'],
    ['/wall', '多镜大屏'],
    ['/mystery', '追星盲盒'],
    ['/gear', '器材 Rig'],
    ['/plans', '拍摄计划'],
  ] as const

  return (
    <div className="min-h-full w-full flex">
      <aside className="w-56 shrink-0 border-r border-white/10 bg-[#0b0e16] p-4 flex flex-col gap-6">
        <div>
          <div className="text-gold font-semibold tracking-[0.2em] text-xs">GOTOSKY</div>
          <h1 className="text-lg font-semibold mt-1">观测星象盾</h1>
          <p className="text-mute text-xs mt-1">Mini 夜空决策台</p>
        </div>
        <nav className="flex flex-col gap-1">
          {nav.map(([to, label]) => (
            <NavLink key={to} to={to} end={to === '/'} className={({ isActive }) => `px-3 py-2 rounded-lg text-sm ${isActive ? 'bg-gold/15 text-gold' : 'text-mute hover:text-ink hover:bg-white/5'}`}>
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto text-xs text-mute space-y-1">
          <div>气象 {meta?.weather_provider || '—'}</div>
          <div>驱动 {meta?.telescope_driver || '—'}</div>
          <div>视宁度：推导值</div>
        </div>
      </aside>
      <main className="flex-1 min-w-0 flex flex-col">
        <header className="h-14 border-b border-white/10 px-6 flex items-center gap-4">
          <label className="text-xs text-mute">站点</label>
          <select aria-label="观测站点" className="bg-panel border border-white/10 rounded-md px-3 py-1.5 text-sm" value={site} onChange={(e) => setSite(e.target.value)}>
            {sites.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>
          <div className="ml-auto text-xs text-mute">
            刷新气象预计 <span className="text-gold">¥0</span> · 配额剩余 {quota?.remaining ?? '—'}/{quota?.daily_limit ?? 2000}
          </div>
        </header>
        <div className="flex-1 p-6">
          <Routes>
            <Route path="/" element={<CalendarPage siteId={site} />} />
            <Route path="/heatmap" element={<HeatmapPage siteId={site} />} />
            <Route path="/sky" element={<SkyPage siteId={site} />} />
            <Route path="/wall" element={<WallPage />} />
            <Route path="/mystery" element={<MysteryPage siteId={site} />} />
            <Route path="/gear" element={<GearPage />} />
            <Route path="/plans" element={<PlansPage siteId={site} />} />
          </Routes>
        </div>
      </main>
      <ToastHost />
    </div>
  )
}
