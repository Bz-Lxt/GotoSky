import { useEffect, useState } from 'react'

type T = { id: number; msg: string }
let seq = 1
const listeners = new Set<(xs: T[]) => void>()
let items: T[] = []

export function pushToast(msg: string) {
  const t = { id: seq++, msg }
  items = [...items, t]
  listeners.forEach((l) => l(items))
  setTimeout(() => dismiss(t.id), 5000)
}

function dismiss(id: number) {
  items = items.filter((x) => x.id !== id)
  listeners.forEach((l) => l(items))
}

export function ToastHost() {
  const [xs, setXs] = useState<T[]>([])
  useEffect(() => {
    listeners.add(setXs)
    return () => { listeners.delete(setXs) }
  }, [])
  return (
    <div className="fixed right-4 bottom-4 space-y-2 z-50">
      {xs.map((t) => (
        <div key={t.id} className="bg-panel border border-gold/30 px-4 py-3 rounded-lg shadow-xl min-w-[240px] flex gap-3">
          <div className="text-sm flex-1">{t.msg}</div>
          <button aria-label="关闭提示" className="text-mute hover:text-ink" onClick={() => dismiss(t.id)}>×</button>
        </div>
      ))}
    </div>
  )
}
