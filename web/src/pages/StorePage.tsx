import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import type { DiscoverBranch } from '../types'

export default function StorePage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()

  useEffect(() => {
    if (!slug) {
      navigate('/', { replace: true })
      return
    }
    let active = true

    api<DiscoverBranch[]>('/discover')
      .then((branches) => {
        if (!active) {
          return
        }
        const match = branches.find((branch) => branch.store_slug === slug)
        navigate(match ? `/branch/${match.branch_id}` : '/', { replace: true })
      })
      .catch(() => {
        if (active) {
          navigate('/', { replace: true })
        }
      })

    return () => {
      active = false
    }
  }, [navigate, slug])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p>Redirecting...</p>
    </div>
  )
}
