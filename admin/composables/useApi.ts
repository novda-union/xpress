interface ApiRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: BodyInit | Record<string, unknown> | unknown[] | null
  headers?: Record<string, string>
}

function isNativeBody(body: ApiRequestOptions['body']): body is BodyInit {
  return typeof body === 'string' ||
    body instanceof FormData ||
    body instanceof Blob ||
    body instanceof URLSearchParams ||
    body instanceof ArrayBuffer ||
    ArrayBuffer.isView(body)
}

function toRequestBody(body: ApiRequestOptions['body']): BodyInit | undefined {
  if (body === undefined || body === null) {
    return undefined
  }

  if (isNativeBody(body)) {
    return body
  }

  return JSON.stringify(body)
}

export function useApi() {
  const config = useRuntimeConfig()
  const { getAuthHeaders } = useAuth()

  async function api<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
    const url = new URL(path, config.public.apiBase).toString()
    const requestBody = toRequestBody(options.body)
    const isJsonBody = requestBody !== undefined && !isNativeBody(options.body)
    const headers = {
      ...getAuthHeaders(),
      ...(isJsonBody ? { 'Content-Type': 'application/json' } : {}),
      ...(options.headers ?? {}),
    }

    const response = await fetch(url, {
      method: options.method ?? 'GET',
      headers,
      body: requestBody,
    })

    if (!response.ok) {
      const message = await response.text()
      throw new Error(message || `Request failed with status ${response.status}`)
    }

    return response.json() as Promise<T>
  }

  return { api }
}
