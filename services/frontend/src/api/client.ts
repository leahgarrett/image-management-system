export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly detail: string,
    public readonly type: string = 'error',
  ) {
    super(detail)
    this.name = 'ApiError'
  }
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  })

  if (!res.ok) {
    let detail = res.statusText
    let type = 'error'
    try {
      const body = await res.json()
      detail = body.detail ?? body.message ?? detail
      type = body.type ?? type
    } catch {
      // non-JSON error body — use statusText
    }
    throw new ApiError(res.status, detail, type)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
