import { useEffect, useState } from 'react'
import { api, fmt } from '../api'
import { pushToast } from '../ui/Toast'

export function PlansPage({ siteId }: { siteId: string }) {
  const [plans, setPlans] = useState<any[]>([])
  useEffect(() => { if (siteId) api.plans(siteId).then(setPlans).catch((e) => pushToast(e.message)) }, [siteId])
  return (
    <div className="space-y-5 w-full">
      <h2 className="text-2xl font-semibold">拍摄计划</h2>
      {plans.map((p) => (
        <article key={p.id} className="bg-panel rounded-2xl border border-white/10 p-5">
          <h3 className="font-semibold">{p.title}</h3>
          <p className="text-mute text-sm">{p.notes}</p>
          <div className="mt-3 space-y-2">
            {(p.items || []).map((it: any) => (
              <div key={it.id} className="flex justify-between text-sm bg-night rounded-lg px-3 py-2">
                <span className="font-mono">{it.target_id.slice(0, 8)}</span>
                <span>{fmt(it.start_utc)} → {fmt(it.end_utc)}</span>
                <span>{it.exposure_s}s × {it.frame_count}</span>
              </div>
            ))}
            {(p.items || []).length === 0 && <div className="text-mute text-sm">还没有计划项，去日历拖入天体。</div>}
          </div>
        </article>
      ))}
    </div>
  )
}
