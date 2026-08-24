import { useEffect, useRef, useState } from 'react'
import { api } from '../api'
import { pushToast } from '../ui/Toast'

export function SkyPage({ siteId }: { siteId: string }) {
  const [targets, setTargets] = useState<any[]>([])
  const [tid, setTid] = useState('')
  const [track, setTrack] = useState<any[]>([])
  const [stars, setStars] = useState<any[]>([])
  const canvas = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    api.targets().then((xs) => { setTargets(xs); const m = xs.find((t) => t.catalog_id === 'M31'); setTid((m || xs[0])?.id) })
    api.stars().then(setStars).catch(() => {})
  }, [])
  useEffect(() => { if (siteId && tid) api.skytrack(siteId, tid).then(setTrack).catch((e) => pushToast(e.message)) }, [siteId, tid])

  useEffect(() => {
    const c = canvas.current
    if (!c) return
    const ctx = c.getContext('2d')!
    let raf = 0
    const draw = () => {
      const w = c.width = c.clientWidth * devicePixelRatio
      const h = c.height = 520 * devicePixelRatio
      ctx.setTransform(devicePixelRatio, 0, 0, devicePixelRatio, 0, 0)
      const W = c.clientWidth, H = 520
      ctx.fillStyle = '#05070d'
      ctx.fillRect(0, 0, W, H)
      const cx = W / 2, cy = H * 0.92, R = Math.min(W, H) * 0.88
      ctx.strokeStyle = 'rgba(255,210,140,0.15)'
      ctx.beginPath(); ctx.arc(cx, cy, R, Math.PI, 0); ctx.stroke()
      for (const alt of [0, 30, 60]) {
        ctx.beginPath(); ctx.arc(cx, cy, R * (1 - alt / 90), Math.PI, 0); ctx.stroke()
      }
      ctx.fillStyle = '#1a140c'
      ctx.fillRect(0, cy, W, H - cy)
      ctx.fillStyle = 'rgba(255,255,255,0.55)'
      for (const s of stars) {
        const p = proj(s.az ?? azOf(s), s.alt ?? 40, cx, cy, R)
        if (p) { ctx.globalAlpha = Math.max(0.2, 1 - (s.mag || 1) / 4); ctx.fillRect(p.x, p.y, 1.5, 1.5) }
      }
      ctx.globalAlpha = 1
      ctx.strokeStyle = '#f5c451'
      ctx.lineWidth = 2
      ctx.beginPath()
      let first = true
      for (const p of track) {
        const pt = proj(p.az, p.alt, cx, cy, R)
        if (!pt) continue
        if (first) { ctx.moveTo(pt.x, pt.y); first = false } else ctx.lineTo(pt.x, pt.y)
      }
      ctx.stroke()
      const moon = track.find((p) => p.moon_alt > 0)
      if (moon) {
        const mp = proj(moon.moon_az, moon.moon_alt, cx, cy, R)
        if (mp) { ctx.fillStyle = '#e8e4d8'; ctx.beginPath(); ctx.arc(mp.x, mp.y, 6, 0, Math.PI * 2); ctx.fill() }
      }
      ctx.fillStyle = '#8b93a7'
      ctx.font = '12px IBM Plex Mono'
      ctx.fillText('N', cx - 4, cy - R - 8)
      ctx.fillText('地平遮挡', 16, H - 16)
      raf = requestAnimationFrame(draw)
    }
    raf = requestAnimationFrame(draw)
    return () => cancelAnimationFrame(raf)
  }, [track, stars])

  return (
    <div className="space-y-4 w-full">
      <div className="flex items-center gap-3">
        <h2 className="text-2xl font-semibold">虚拟夜空对焦星图</h2>
        <select aria-label="星图目标" className="bg-panel border border-white/10 rounded-md px-3 py-1.5 text-sm" value={tid} onChange={(e) => setTid(e.target.value)}>
          {targets.slice(0, 80).map((t) => <option key={t.id} value={t.id}>{t.catalog_id} {t.name_zh}</option>)}
        </select>
      </div>
      <canvas ref={canvas} className="w-full rounded-2xl border border-white/10" data-testid="sky-canvas" style={{ height: 520 }} />
      <p className="text-mute text-sm">金色曲线为目标周日轨迹；白点为月亮。Alt/Az 由后端 Meeus 引擎计算。</p>
    </div>
  )
}

function azOf(s: any) { return ((s.ra_hours || 0) * 15) % 360 }
function proj(az: number, alt: number, cx: number, cy: number, R: number) {
  if (alt < 0) return null
  const r = R * (1 - alt / 90)
  const a = (az - 180) * Math.PI / 180
  return { x: cx + r * Math.sin(a), y: cy - r * Math.cos(a) }
}
