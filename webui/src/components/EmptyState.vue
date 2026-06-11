<template>
<tr>
    <td :colspan="colspan" class="text-center py-5 text-body-secondary">
        <CIcon :icon="icon" size="3xl" class="mb-3 opacity-50 d-block mx-auto" />
        <p class="mb-3">{{ message }}</p>
        <CButton v-if="actionLabel" color="primary" size="sm" @click="emit('action-clicked')">
            <CIcon icon="cilPlus" class="me-1" />{{ actionLabel }}
        </CButton>
    </td>
</tr>
</template>

<script setup>
defineProps({
    icon: { type: String, default: 'cilInbox' },
    message: { type: String, default: 'No items found.' },
    actionLabel: { type: String, default: '' },
    colspan: { type: Number, default: 99 },
})
const emit = defineEmits(['action-clicked'])
</script>
