<template>
<CForm @submit.prevent="createPrAddr">
    <div class="mb-3">
        <CFormLabel for="praddr-email">Email Address</CFormLabel>
        <CFormInput id="praddr-email" v-model="praddr_email" type="email" placeholder="you@example.com" />
    </div>
    <div class="mb-3">
        <CFormLabel for="comment">Comment</CFormLabel>
        <CFormInput id="comment" v-model="comment" placeholder="Optional note" />
    </div>
    <CAlert v-if="errorMessage" color="danger" class="mb-3">{{ errorMessage }}</CAlert>
    <div class="d-flex gap-2">
        <CButton type="submit" color="primary" :disabled="submitting">
            <CSpinner v-if="submitting" size="sm" class="me-1" />Create
        </CButton>
        <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
    </div>
</CForm>
</template>

<script setup>
import { ref } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['done', 'created'])

const praddr_email = ref('')
const comment = ref('')
const errorMessage = ref('')
const submitting = ref(false)

const createPrAddr = async () => {
    errorMessage.value = ''
    submitting.value = true
    const res = await apiFetch('/api/v1/praddrs', {
        method: 'POST',
        body: JSON.stringify({
            email: praddr_email.value,
            metadata: { comment: comment.value },
        }),
    })
    const json = await res.json()
    submitting.value = false
    if (res.status === 201) {
        emit('created', json.email)
    } else {
        errorMessage.value = json.errors?.[0]?.detail ?? 'An unexpected error occurred'
    }
}
</script>
