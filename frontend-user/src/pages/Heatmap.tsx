import { useEffect, useMemo, useState } from 'react'
import { api, tierLabel } from '../api'
import { pushToast } from '../ui/Toast'

const FACTORS = [
  ['score', '总分'],
  ['factor_c', '云量'],
  ['factor_s', '视宁度'],
  ['factor_m', '月相'],
  ['factor_a', '高度'],
  ['factor_t', '透明度'],
  ['factor_l', '光害'],
  ['factor_n', '夜深'],
] as const

export function HeatmapPage({ siteId }: { siteId: string }) {
  const [targets, setTargets] = useState<any[]>([])
  const [tid, setTid] = useState('')
  const [slots, setSlots] = useState<any[]>([])
  const [layer, setLayer] = useState<string>('score')

  useEffect(() => { api.targets().then((xs) => { setTargets(xs); const m = xs.find((t) => t.catalog_id === 'M31'); setTid((m || xs[0])?.id) }) }, [])
  useEffect(() => { if (siteId && tid) api.heatmap(siteId, tid).then(setSlots).catch((e) => pushToast(e.message)) }, [siteId, tid])

  const grid = useMemo(() => {
    const days: Record<string, any[]> = {}
    for (const s of slots) {
      const d = new Date(s.slot_utc)
      const key = `${d.getMonth()+1}/${d.getDate()}`
      days[key] = days[key] || []
      days[key].push(s)
    }
    return Object.entries(days).slice(0, 7)
  }, [slots])

  function color(v: number, isScore: boolean) {
    const n = isScore ? v / 100 : v
    if (n <= 0) return '#1a1e27'
    if (n >= 0.8) return '#f5c451'
    if (n >= 0.65) return '#c9d4e0'
    if (n >= 0.5) return '#c9844a'
    return '#2d3a55'
  }

  return (
    <div className="space-y-5 w-full">
      <div className="flex items-center gap-3 flex-wrap">
        <h2 className="text-2xl font-semibold">7×24 天气热力矩阵</h2>
        <select aria-label="热力目标" className="bg-panel border border-white/10 rounded-md px-3 py-1.5 text-sm" value={tid} onChange={(e) => setTid(e.target.value)}>
          {targets.slice(0, 60).map((t) => <option key={t.id} value={t.id}>{t.catalog_id}</option>)}
        </select>
        <div className="flex gap-1 flex-wrap">
          {FACTORS.map(([k, lab]) => (
            <button key={k} onClick={() => setLayer(k)} className={`px-2 py-1 rounded-md text-xs ${layer === k ? 'bg-gold text-night' : 'bg-white/5 text-mute'}`}>{lab}</button>
          ))}
        </div>
      </div>
      <p className="text-mute text-sm">视宁度层为高空急流推导值，不是观测值。数据来自后端子分，前端不重算。</p>
      <div className="bg-panel rounded-2xl border border-white/10 p-4 overflow-x-auto" data-testid="heatmap">
        {grid.length === 0 && <div className="text-mute p-8">暂无得分。请到日历页刷新气象。</div>}
        {grid.map(([day, xs]) => (
          <div key={day} className="flex items-center gap-2 mb-2">
            <div className="w-12 text-xs text-mute">{day}</div>
            <div className="flex gap-1">
              {Array.from({ length: 24 }, (_, h) => {
                const s = xs.find((x) => new Date(x.slot_utc).getHours() === h)
                const val = s ? (layer === 'score' ? s.score : s[layer] * (layer === 'score' ? 1 : 100)) : 0
                return <div key={h} title={`${h}:00 ${s ? (layer==='score'?s.score:s[layer]) : '—'}`} className="w-7 h-7 rounded-sm" style={{ background: s ? color(layer==='score'?s.score:s[layer], layer==='score') : '#12151c' }} />
              })}
            </div>
          </div>
        ))}
      </div>
      <div className="text-xs text-mute">{slots[0] ? `样本 ${tierLabel(slots[0].tier)} · 限制 ${slots[0].limiting_factor} · seeing ${slots[0].seeing_arcsec?.toFixed?.(2)}″ (derived)` : ''}</div>
    </div>
  )
}
