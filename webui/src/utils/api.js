let _refreshPromise = null

async function tryRefreshToken() {
  if (!_refreshPromise) {
    const provider = sessionStorage.getItem('oidcProvider')
    if (!provider) return false

    _refreshPromise = fetch(`/auth/${provider}/refresh`, { credentials: 'include' })
      .then(res => res.ok)
      .catch(() => false)
      .finally(() => { _refreshPromise = null })
  }
  return _refreshPromise
}

export async function apiFetch(url, options = {}) {
  const method = (options.method || 'GET').toUpperCase()
  const headers = method !== 'GET'
    ? { 'Content-Type': 'application/json', ...options.headers }
    : options.headers

  const res = await fetch(url, { ...options, headers, credentials: 'include' })

  if (res.status === 401) {
    const refreshed = await tryRefreshToken()
    if (refreshed) {
      return fetch(url, { ...options, headers, credentials: 'include' })
    }
    sessionStorage.removeItem('oidcProvider')
    window.location.href = '/'
    return
  }

  return res
}
