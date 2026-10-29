<template>
<div class="d-flex align-items-center gap-3">
    <div class="text-end lh-sm">
        <div class="small text-body-secondary">
            Logged as: <strong>{{ props.userInfo.first_name }} {{ props.userInfo.last_name }}</strong>
        </div>
        <div class="text-body-secondary" style="font-size: 0.75rem;">{{ props.userInfo.login }}</div>
    </div>
    <CButton color="danger" variant="outline" size="sm" :disabled="loggingOut" @click="logout">
        Logout
    </CButton>
</div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps(['userInfo'])
const loggingOut = ref(false)

// logout is a state-changing request, so it must be a POST
const logout = async () => {
    loggingOut.value = true
    try {
        await fetch('/auth/logout', { method: 'POST', credentials: 'include' })
    } finally {
        sessionStorage.removeItem('oidcProvider')
        window.location.href = '/'
    }
}
</script>
