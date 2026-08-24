import { useEffect, useState } from 'react'
import { api } from '../api'
import { pushToast } from '../ui/Toast'

export function MysteryPage({ siteId }: { siteId: string }) {
  const [items, setItems] = useState<any[]>([])
  useEffect(() => { if (siteId) api.mystery(siteId).then(setItems).catch((e) => pushToast(e.message)) }, [siteId])
  return (
    <div className="space-y-5 w-full">
      <h2 className="text-2xl font-semibold">自动追星盲盒</h2>
      <p className="text-mute text-sm">这不是随机抽取。引擎按当夜均分 × 视场匹配给出 1–3 个可解释推荐。</p>
      <div className="grid grid-cols-3 gap-4" data-testid="mystery-box">
        {items.map((c, i) => (
          <article key={i} className="bg-panel rounded-2xl border border-gold/20 p-5">
            <div className="text-gold text-xs tracking-widest">BOX {i + 1}</div>
            <h3 className="text-xl mt-2">{c.Target?.catalog_id || c.target?.catalog_id} {c.Target?.name_zh || c.target?.name_zh}</h3>
            <div className="mt-3 text-sm">均分 {c.MeanScore?.toFixed?.(0) ?? c.mean_score} · 峰值 {c.PeakScore ?? c.peak_score}</div>
            <p className="mt-3 text-mute text-sm">{c.Reason || c.reason}</p>
          </article>
        ))}
        {items.length === 0 && <div className="text-mute col-span-3">暂无推荐。请先刷新气象生成黄金窗口。</div>}
      </div>
    </div>
  )
}
