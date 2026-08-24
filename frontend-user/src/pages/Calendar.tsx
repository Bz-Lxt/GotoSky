import { useEffect, useMemo, useState } from 'react'
import { api, fmt, tierLabel } from '../api'
import { pushToast } from '../ui/Toast'

export function CalendarPage({ siteId }: { siteId: string }) {
  const [targets, setTargets] = useState<any[]>([])
  const [scores, setScores] = useState<any[]>([])
  const [windows, setWindows] = useState<any[]>([])
  const [tid, setTid] = useState('')
  const [planId, setPlanId] = useState('')
  const [q, setQ] = useState('')

  useEffect(() => {
    api.targets().then((xs) => { setTargets(xs); const m31 = xs.find((t) => t.catalog_id === 'M31'); if (m31) setTid(m31.id) }).catch((e) => pushToast(e.message))
    api.plans(siteId).then((ps) => { if (ps[0]) setPlanId(ps[0].id) }).catch(() => {})
  }, [siteId])

  useEffect(() => {
    if (!siteId || !tid) return
    api.scores(siteId, tid).then(setScores).catch((e) => pushToast(e.message))
    api.windows(siteId).then(setWindows).catch(() => {})
  }, [siteId, tid])

  const night = useMemo(() => {
    const now = new Date()
    const start = new Date(now)
    start.setHours(16, 0, 0, 0)
    return Array.from({ length: 16 }, (_, i) => {
      const h = new Date(start.getTime() + i * 3600_000)
      const slot = scores.find((s) => new Date(s.slot_utc).getHours() === h.getHours() && Math.abs(new Date(s.slot_utc).getTime() - h.getTime()) < 36e5 * 20)
      return { h, slot }
    })
  }, [scores])

  const filtered = targets.filter((t) => (t.catalog_id + t.name + t.name_zh).toLowerCase().includes(q.toLowerCase()))

  async function dropTarget(targetId: string) {
    if (!planId) {
      const p: any = await api.createPlan({ site_id: siteId, title: '拖入计划', notes: '' })
      setPlanId(p.id)
      await add(p.id, targetId)
      return
    }
    await add(planId, targetId)
  }

  async function add(pid: string, targetId: string) {
    const start = new Date()
    start.setMinutes(0, 0, 0)
    const end = new Date(start.getTime() + 2 * 3600_000)
    await api.addItem(pid, { target_id: targetId, start_utc: start.toISOString(), end_utc: end.toISOString(), exposure_s: 180, frame_count: 20, filter_seq: [], narrowband: false })
    pushToast('已将天体拖入拍摄计划')
  }

  async function refresh() {
    try {
      await api.refresh(siteId)
      pushToast('气象已刷新 · ¥0')
      if (tid) setScores(await api.scores(siteId, tid))
    } catch (e: any) { pushToast(e.message) }
  }

  return (
    <div className="space-y-6 w-full">
      <div className="flex items-end justify-between gap-4 flex-wrap">
        <div>
          <h2 className="text-2xl font-semibold">24 小时夜间日历</h2>
          <p className="text-mute text-sm mt-1">金 / 银 / 铜由后端星空纯净度得分自动着色。视宁度为推导值。</p>
        </div>
        <button onClick={refresh} className="px-4 py-2 rounded-lg bg-gold text-night text-sm font-medium" data-testid="refresh-weather">立即刷新气象（¥0）</button>
      </div>
      <div className="grid grid-cols-12 gap-6">
        <section className="col-span-9 bg-panel rounded-2xl border border-white/10 p-5">
          <div className="flex items-center gap-3 mb-4">
            <select aria-label="目标天体" className="bg-night border border-white/10 rounded-md px-3 py-1.5 text-sm" value={tid} onChange={(e) => setTid(e.target.value)}>
              {targets.slice(0, 80).map((t) => <option key={t.id} value={t.id}>{t.catalog_id} {t.name_zh || t.name}</option>)}
            </select>
          </div>
          <div className="grid grid-cols-8 gap-2" data-testid="night-strip">
            {night.map(({ h, slot }) => (
              <div key={h.toISOString()} className={`rounded-xl p-3 min-h-[92px] ${slot ? 'tier-' + slot.tier : 'bg-white/5'}`}>
                <div className="font-mono text-xs opacity-80">{String(h.getHours()).padStart(2, '0')}:00</div>
                <div className="text-lg font-semibold mt-1">{slot ? slot.score : '—'}</div>
                <div className="text-[11px] mt-1">{slot ? tierLabel(slot.tier) : '无数据'}</div>
                {slot?.limiting_factor && <div className="text-[10px] mt-1 opacity-80">{slot.limiting_factor}</div>}
              </div>
            ))}
          </div>
          <div className="mt-6">
            <h3 className="text-sm text-mute mb-2">黄金窗口 Top</h3>
            <div className="space-y-2">
              {windows.filter((w) => !tid || w.target_id === tid).slice(0, 3).map((w) => (
                <div key={w.id} className="flex items-center justify-between bg-night/60 rounded-lg px-3 py-2 text-sm">
                  <span className={`px-2 py-0.5 rounded-md text-xs tier-${w.tier}`}>{tierLabel(w.tier)}</span>
                  <span className="font-mono">{fmt(w.start_local || w.start_utc)} → {fmt(w.end_local || w.end_utc)}</span>
                  <span>均分 {w.mean_score?.toFixed?.(0)} · 限制 {w.limiting_factor}</span>
                </div>
              ))}
              {windows.length === 0 && <div className="text-mute text-sm">暂无窗口，请先刷新气象。</div>}
            </div>
          </div>
        </section>
        <aside className="col-span-3 bg-panel rounded-2xl border border-white/10 p-4">
          <h3 className="font-medium mb-2">星表 · 拖入计划</h3>
          <input className="w-full bg-night border border-white/10 rounded-md px-3 py-1.5 text-sm mb-3" placeholder="搜索 M31 / 仙女" value={q} onChange={(e) => setQ(e.target.value)} />
          <div className="space-y-1 max-h-[560px] overflow-auto" data-testid="target-list">
            {filtered.slice(0, 40).map((t) => (
              <div key={t.id} draggable onDragStart={(e) => e.dataTransfer.setData('text/target', t.id)}
                onDoubleClick={() => dropTarget(t.id)}
                className="px-3 py-2 rounded-lg hover:bg-white/5 cursor-grab text-sm flex justify-between">
                <span>{t.catalog_id}</span>
                <span className="text-mute">{t.name_zh || t.name}</span>
              </div>
            ))}
          </div>
          <div className="mt-4 p-3 rounded-xl border border-dashed border-gold/30 text-xs text-mute" data-testid="drop-plan"
            onDragOver={(e) => e.preventDefault()}
            onDrop={(e) => { e.preventDefault(); const id = e.dataTransfer.getData('text/target'); if (id) dropTarget(id) }}>
            将天体拖到此处，或双击加入计划
          </div>
        </aside>
      </div>
    </div>
  )
}
