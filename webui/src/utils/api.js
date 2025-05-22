export async function apiFetch(url, options = {}) {

  const res = await fetch(url, {
    ...options,
    credentials: 'include'
  })

  if (res.status === 401) {
    window.location.reload()
    return
  }

  return res
}
