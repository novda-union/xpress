interface BranchConflictSheetProps {
  newBranchName: string
  onConfirm: () => void
  onCancel: () => void
}

export function BranchConflictSheet({ newBranchName, onConfirm, onCancel }: BranchConflictSheetProps) {
  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        aria-label="Cancel"
        onClick={onCancel}
        className="absolute inset-0 bg-[var(--xp-overlay)]"
      />
      <div className="absolute inset-x-0 bottom-0 rounded-t-[20px] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-5">
        <div className="mx-auto mb-5 h-1 w-8 rounded-full bg-[var(--tg-theme-hint-color)]/40" />
        <h2 className="text-[18px] font-bold">Add from {newBranchName}?</h2>
        <p className="mt-2 text-[14px] leading-6 text-[var(--tg-theme-hint-color)]">
          You already have items from another branch. They'll stay in a separate cart — you can place each order independently.
        </p>
        <div className="mt-6 flex flex-col gap-3 pb-4">
          <button
            type="button"
            onClick={onConfirm}
            className="flex h-[52px] w-full items-center justify-center rounded-[20px] bg-[var(--xp-brand)] text-sm font-semibold text-white"
          >
            Add to new cart
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="flex h-[52px] w-full items-center justify-center rounded-[20px] border border-[var(--xp-border)] text-sm font-semibold"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}
