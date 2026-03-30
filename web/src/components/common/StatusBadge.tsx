const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-slate-400/15 text-slate-500',
  accepted: 'bg-blue-500/15 text-blue-600',
  preparing: 'bg-amber-500/15 text-amber-600',
  ready: 'bg-green-500/15 text-green-600',
  picked_up: 'bg-indigo-500/15 text-indigo-600',
  rejected: 'bg-red-500/15 text-red-600',
  cancelled: 'bg-slate-400/15 text-slate-500',
}

const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  accepted: 'Accepted',
  preparing: 'Preparing',
  ready: 'Ready',
  picked_up: 'Picked Up',
  rejected: 'Rejected',
  cancelled: 'Cancelled',
}

export function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`inline-flex rounded-full px-4 py-2 text-[15px] font-semibold ${STATUS_COLORS[status] ?? STATUS_COLORS.pending}`}>
      {STATUS_LABELS[status] ?? status}
    </span>
  )
}
