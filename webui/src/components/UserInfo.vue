<template>
    <div class="d-flex align-items-center gap-3">
        <span class="text-body-secondary small">
            {{ props.userInfo.first_name }} {{ props.userInfo.last_name }}
            <CBadge color="secondary" class="ms-1">{{ props.userInfo.type }}</CBadge>
        </span>
        <CButton component="a" href="/auth/logout" color="danger" variant="outline" size="sm">
            Logout
        </CButton>
    </div>
</template>

<script setup>
const props = defineProps(['userInfo'])
</script>
