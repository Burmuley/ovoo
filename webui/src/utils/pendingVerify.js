const STORAGE_KEY = 'ovoo_pending_praddr_verify'

export function setPendingVerify(id, token) {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ id, token }))
}

export function getPendingVerify() {
  const raw = sessionStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function clearPendingVerify() {
  sessionStorage.removeItem(STORAGE_KEY)
}
