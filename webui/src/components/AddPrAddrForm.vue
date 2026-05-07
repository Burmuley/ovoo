<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New Protected Address</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createPrAddr">
                <div class="mb-3">
                    <CFormLabel for="praddr-email">Email Address</CFormLabel>
                    <CFormInput id="praddr-email" v-model="praddr_email" type="email" placeholder="you@example.com" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="comment">Comment</CFormLabel>
                    <CFormInput id="comment" v-model="comment" placeholder="Optional note" />
                </div>
                <div class="d-flex gap-2">
                    <CButton type="submit" color="primary">Create</CButton>
                    <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
                </div>
            </CForm>
            <CAlert v-if="result.status === 201" color="success" class="mt-3">
                Protected address <strong>{{ result.json.email }}</strong> was successfully created.
            </CAlert>
            <CAlert v-else-if="result.status" color="danger" class="mt-3">
                <div v-for="error in result.json.errors" :key="error.detail">{{ error.detail }}</div>
            </CAlert>
        </CCardBody>
    </CCard>
</template>

<script setup>
import { ref } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['done'])

const praddr_email = ref('')
const comment = ref('')
const result = ref({})

const createPrAddr = async () => {
    const res = await apiFetch('/api/v1/praddrs', {
        method: 'POST',
        body: JSON.stringify({
            email: praddr_email.value,
            metadata: { comment: comment.value },
        }),
    })
    result.value = { status: res.status, json: await res.json() }
}
</script>
