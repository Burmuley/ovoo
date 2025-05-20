<template>
    <div v-if="!isAuthenticated">
        <Login />
    </div>
    <div v-else>
        <MainTabs />
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Login from './components/Login.vue'
import MainTabs from './components/MainTabs.vue'

const isAuthenticated = ref(false)

function hasAuthCookie() {
    return document.cookie.split('; ').some(row => row.startsWith('ovoo_auth='))
}

onMounted(() => {
    isAuthenticated.value = hasAuthCookie()
})
</script>
