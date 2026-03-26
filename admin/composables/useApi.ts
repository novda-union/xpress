export function useApi() {
  const config = useRuntimeConfig()
  const { getAuthHeaders } = useAuth()

  async function api<T>(path: string, options: any = {}): Promise<T> {
    return $fetch<T>(`${config.public.apiBase}${path}`, {
      ...options,
      headers: {
        ...getAuthHeaders(),
        ...options.headers,
      },
    })
  }

  return { api }
}
