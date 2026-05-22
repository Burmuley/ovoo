<template>
    <div v-if="authChecked && !isAuthenticated">
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

const isAuthenticated = ref(false)
const authChecked = ref(false)

onMounted(async () => {
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
})
</script>
