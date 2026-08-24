import { useEffect, useState } from 'react'
import { api } from '../api'
import { pushToast } from '../ui/Toast'
import { Dialog } from '../ui/Dialog'

export function WallPage() {
  const [rigs, setRigs] = useState<any[]>([])
  const [sessions, setSessions] = useState<any[]>([])
  const [live, setLive] = useState<any>({})
  const [confirm, setConfirm] = useState<string | null>(null)

  useEffect(() => {
    api.rigs().then(setRigs).catch((e) => pushToast(e.message))
    api.sessions().then(setSessions).catch(() => {})
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${location.host}/ws`)
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data)
      if (msg.type === 'telemetry') {
        setLive((p: any) => ({ ...p, [msg.data.session_id]: msg.data }))
      }
    }
    const t = setInterval(() => api.sessions().then(setSessions).catch(() => {}), 4000)
    return () => { ws.close(); clearInterval(t) }
  }, [])

  async function start(rig: any) {
    try {
      await api.startSession({ rig_id: rig.id, ra: 0.712, dec: 41.27, frames: 6, exposure_s: 1 })
      pushToast('会话已启动（Mock 驱动）')
      setSessions(await api.sessions())
    } catch (e: any) { pushToast(e.message) }
  }

  return (
    <div className="space-y-5 w-full">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">多镜协同虚拟大屏</h2>
        <span className="text-xs text-mute">来源由运行模式标记 · 默认 SIMULATED</span>
      </div>
      <div className="grid grid-cols-2 gap-4" data-testid="rig-wall">
        {rigs.map((r) => {
          const sess = sessions.find((s) => s.rig_id === r.id) || sessions[0]
          const tel = sess ? live[sess.id] : null
          const st = tel?.state || sess?.state || 'IDLE'
          return (
            <article key={r.id} className="bg-panel rounded-2xl border border-white/10 p-5">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-semibold">{r.name}</h3>
                <span className="text-xs px-2 py-0.5 rounded-md bg-gold/15 text-gold">{st}</span>
              </div>
              <div className="grid grid-cols-3 gap-3 text-sm">
                {[['赤道仪 RA', tel?.sensors?.mount_ra ?? '—'], ['Dec', tel?.sensors?.mount_dec ?? '—'], ['导星 RMS', tel?.sensors?.guide_rms ?? '—'],
                  ['CCD °C', tel?.sensors?.ccd_temp ?? '—'], ['滤镜位', tel?.sensors?.filter_pos ?? '—'], ['湿度', tel?.sensors?.humidity ?? '—']].map(([k, v]) => (
                  <div key={k} className="bg-night rounded-xl p-3">
                    <div className="text-mute text-xs">{k}</div>
                    <div className="font-mono text-lg mt-1">{typeof v === 'number' ? v.toFixed(2) : v}</div>
                  </div>
                ))}
              </div>
              <div className="mt-4 text-xs text-mute">进度 {tel?.progress_k ?? sess?.progress_k ?? 0}/{tel?.progress_n ?? sess?.progress_n ?? 0} · 来源 {tel?.source || sess?.source_mode || 'SIMULATED'}</div>
              <div className="mt-4 flex gap-2">
                <button className="px-3 py-1.5 rounded-md bg-gold text-night text-sm" onClick={() => start(r)}>启动追星</button>
                {sess && <button className="px-3 py-1.5 rounded-md bg-white/5 text-sm" onClick={() => setConfirm(sess.id)}>中止</button>}
              </div>
            </article>
          )
        })}
        {rigs.length === 0 && <div className="text-mute">暂无 Rig，请先看器材页。</div>}
      </div>
      <Dialog open={!!confirm} title="中止拍摄会话？" danger onClose={() => setConfirm(null)} onOk={async () => {
        if (confirm) await api.cmd(confirm, 'ABORT')
        setConfirm(null)
        pushToast('已发送中止')
      }}>户外设备将进入安全停靠，不会重试。</Dialog>
    </div>
  )
}
