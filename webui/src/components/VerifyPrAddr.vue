<template>
<div class="bg-body-tertiary min-vh-100 d-flex align-items-center justify-content-center">
    <CCard style="width: 26rem;">
        <CCardBody class="p-4">
            <h4 class="text-center mb-1">Ovoo</h4>
            <p class="text-center text-body-secondary mb-4">Verify Protected Address</p>

            <div v-if="state === 'checking'" class="text-center py-3">
                <CSpinner />
            </div>

            <template v-else-if="state === 'needsLogin'">
                <p class="text-body-secondary">Sign in to finish verifying your protected address.</p>
                <div class="d-grid gap-2">
                    <CButton v-for="provider in providers" :key="provider" color="primary" variant="outline"
                        @click="loginWith(provider)">
                        Sign in with {{ provider }}
                    </CButton>
                </div>
            </template>

            <template v-else-if="state === 'success'">
                <CAlert color="success" class="mb-3">
                    <strong>{{ verifiedEmail }}</strong> has been verified.
                </CAlert>
                <div class="d-grid">
                    <CButton color="primary" @click="goHome">Go to Ovoo</CButton>
                </div>
            </template>

            <template v-else-if="state === 'error'">
                <CAlert color="warning" class="mb-3" style="word-break: break-word;">
                    {{ errorMessage }}
                </CAlert>
                <div class="d-grid">
                    <CButton color="primary" variant="outline" @click="goHome">Go to Ovoo</CButton>
                </div>
            </template>

            <template v-else-if="state === 'missingToken'">
                <CAlert color="warning" class="mb-3">
                    This verification link looks incomplete. Please use the full link from your email.
                </CAlert>
                <div class="d-grid">
                    <CButton color="primary" variant="outline" @click="goHome">Go to Ovoo</CButton>
                </div>
            </template>
        </CCardBody>
    </CCard>
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import { setPendingVerify, clearPendingVerify } from '../utils/pendingVerify'

const state = ref('checking')
const errorMessage = ref('')
const verifiedEmail = ref('')
const providers = ref([])

const idMatch = window.location.pathname.match(/^\/verify\/([^/]+)/)
const id = idMatch ? idMatch[1] : ''
const token = window.location.hash.startsWith('#') ? window.location.hash.slice(1) : ''

function goHome() {
    window.location.href = '/'
}

function loginWith(provider) {
    setPendingVerify(id, token)
    sessionStorage.setItem('oidcProvider', provider)
    window.location.href = `/auth/${provider}/login`
}

async function loadProviders() {
    const res = await apiFetch('/auth/providers')
    providers.value = await res.json()
}

async function attemptVerify() {
    const res = await apiFetch(`/api/v1/praddrs/${id}/tokenvalidate`, {
        method: 'POST',
        body: JSON.stringify({ token }),
    })
    const data = await res.json()
    clearPendingVerify()
    if (res.status === 201) {
        verifiedEmail.value = data.email
        state.value = 'success'
    } else {
        errorMessage.value = data.errors?.[0]?.detail ?? 'Verification failed. The link may be invalid or expired.'
        state.value = 'error'
    }
}

onMounted(async () => {
    if (!id || !token) {
        state.value = 'missingToken'
        return
    }

    let authenticated = false
    try {
        // Raw fetch (not apiFetch): apiFetch reloads to "/" on 401, which would
        // tear down this page before an unauthenticated visitor could sign in.
        const res = await fetch('/api/v1/users/profile', { credentials: 'include' })
        authenticated = res.ok
    } catch {
        authenticated = false
    }

    if (authenticated) {
        await attemptVerify()
    } else {
        setPendingVerify(id, token)
        await loadProviders()
        state.value = 'needsLogin'
    }
})
</script>
