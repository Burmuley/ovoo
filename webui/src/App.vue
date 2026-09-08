<template>
<div v-if="isVerifyRoute">
    <VerifyPrAddr />
</div>
<div v-else-if="authChecked && !isAuthenticated">
    <Login />
</div>
<div v-else-if="authChecked && isAuthenticated">
    <MainTabs />
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Login from './components/Login.vue'
import MainTabs from './components/MainTabs.vue'
import VerifyPrAddr from './components/VerifyPrAddr.vue'
import { getPendingVerify } from './utils/pendingVerify'

const isAuthenticated = ref(false)
const authChecked = ref(false)
const isVerifyRoute = window.location.pathname.startsWith('/verify/')

onMounted(async () => {
    if (isVerifyRoute) return

    try {
        // Raw fetch (not apiFetch): apiFetch reloads on 401, causing an
        // infinite loop when the user is not logged in.
        const res = await fetch('/api/v1/users/profile', { credentials: 'include' })
        isAuthenticated.value = res.ok
    } catch {
        isAuthenticated.value = false
    } finally {
        authChecked.value = true
    }

    // If the user just signed in to finish an in-progress protected address
    // verification (see VerifyPrAddr.vue), hand them back to that page now
    // that they have a session.
    if (isAuthenticated.value) {
        const pending = getPendingVerify()
        if (pending) {
            window.location.href = `/verify/${pending.id}#${pending.token}`
        }
    }
})
</script>
