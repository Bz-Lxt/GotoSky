export function Dialog({ open, title, children, onClose, onOk, danger }: { open: boolean; title: string; children: any; onClose: () => void; onOk?: () => void; danger?: boolean }) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-40 bg-black/60 flex items-center justify-center p-6" onClick={onClose}>
      <div className="bg-panel border border-white/10 rounded-2xl p-6 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-semibold mb-3">{title}</h3>
        <div className="text-sm text-mute mb-5">{children}</div>
        <div className="flex justify-end gap-2">
          <button className="px-3 py-1.5 rounded-md bg-white/5" onClick={onClose}>取消</button>
          {onOk && <button className={`px-3 py-1.5 rounded-md ${danger ? 'bg-danger text-white' : 'bg-gold text-night'}`} onClick={onOk}>确认</button>}
        </div>
      </div>
    </div>
  )
}
