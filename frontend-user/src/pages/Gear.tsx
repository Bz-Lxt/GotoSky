import { useEffect, useState } from 'react'
import { api } from '../api'
import { pushToast } from '../ui/Toast'

export function GearPage() {
  const [eq, setEq] = useState<any[]>([])
  const [rigs, setRigs] = useState<any[]>([])
  useEffect(() => {
    api.equipment().then(setEq).catch((e) => pushToast(e.message))
    api.rigs().then(setRigs).catch((e) => pushToast(e.message))
  }, [])
  const kinds = ['MOUNT', 'OTA', 'CAMERA', 'FILTER_WHEEL', 'GUIDE_SCOPE', 'GUIDE_CAMERA']
  return (
    <div className="space-y-6 w-full">
      <h2 className="text-2xl font-semibold">器材与多镜 Rig</h2>
      <div className="grid grid-cols-3 gap-3">
        {kinds.map((k) => (
          <section key={k} className="bg-panel rounded-2xl border border-white/10 p-4">
            <h3 className="text-xs text-gold tracking-widest mb-2">{k}</h3>
            {eq.filter((e) => e.kind === k).map((e) => (
              <div key={e.id} className="text-sm py-1 border-b border-white/5 last:border-0">{e.name}</div>
            ))}
            {eq.filter((e) => e.kind === k).length === 0 && <div className="text-mute text-sm">空</div>}
          </section>
        ))}
      </div>
      <div className="bg-panel rounded-2xl border border-white/10 p-5">
        <h3 className="font-medium mb-3">已绑定 Rig</h3>
        {rigs.map((r) => (
          <div key={r.id} className="mb-3">
            <div className="font-semibold">{r.name}</div>
            <div className="text-xs text-mute">{(r.components || []).map((c: any) => `${c.role}:${c.name}`).join(' · ')}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
