export async function apiFetch(url, options = {}) {
    const method = (options.method || 'GET').toUpperCase()
    const headers = method !== 'GET'
        ? { 'Content-Type': 'application/json', ...options.headers }
        : options.headers

    const res = await fetch(url, { ...options, headers, credentials: 'include' })

    if (res.status === 401) {
        window.location.reload()
        return
    }

    return res
}
