<template>
    <div class="d-flex align-items-center gap-3">
        <div class="text-end lh-sm">
            <div class="small text-body-secondary">
                Logged as: <strong>{{ props.userInfo.first_name }} {{ props.userInfo.last_name }}</strong>
            </div>
            <div class="text-body-secondary" style="font-size: 0.75rem;">{{ props.userInfo.login }}</div>
        </div>
        <CButton component="a" href="/auth/logout" color="danger" variant="outline" size="sm">
            Logout
        </CButton>
    </div>
</template>

<script setup>
const props = defineProps(['userInfo'])
</script>
