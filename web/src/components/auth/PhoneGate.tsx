import { LoaderCircle, PhoneCall } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface PhoneGateProps {
  loading: boolean
  error: string
  onRequestAccess: () => Promise<unknown> | void
}

export function PhoneGate({ loading, error, onRequestAccess }: PhoneGateProps) {
  return (
    <div className="min-h-screen flex items-center justify-center xp-page-padding">
      <div className="w-full max-w-sm text-center">
        <div className="mx-auto mb-6 flex h-14 w-14 items-center justify-center rounded-2xl bg-[var(--xp-brand-muted)] text-[var(--xp-brand)]">
          <PhoneCall className="h-7 w-7" />
        </div>
        <h1 className="text-[22px] font-bold">Order without the wait</h1>
        <p className="mt-3 text-[15px] leading-6 text-[var(--tg-theme-hint-color)]">
          We need your phone to identify you and keep your orders safe.
        </p>

        <Button
          type="button"
          onClick={() => void onRequestAccess()}
          disabled={loading}
          className="mt-8 h-14 w-full rounded-[20px] text-base"
        >
          {loading ? <LoaderCircle className="h-5 w-5 animate-spin" /> : <PhoneCall className="h-5 w-5" />}
          Share Phone Number
        </Button>

        <p className="mt-4 text-sm text-[var(--tg-theme-hint-color)]">
          Your number is only used for order tracking.
        </p>

        {error ? (
          <div className="mt-6 rounded-2xl border border-red-200 bg-red-50/90 px-4 py-3 text-left text-sm text-red-700">
            <p className="font-medium">Phone number is required to continue</p>
            <p className="mt-1 opacity-80">{error}</p>
          </div>
        ) : null}
      </div>
    </div>
  )
}
