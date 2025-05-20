export async function fetchWithAuth(url, options = {}) {
  try {
    const response = await fetch(url, {
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      ...options
    })
    if (response.status === 401) {
      window.dispatchEvent(new CustomEvent('unauthorized'))
      throw new Error('Unauthorized')
    }
    return response
  } catch (err) {
    console.error('Fetch error:', err)
    throw err
  }
}
