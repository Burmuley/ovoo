<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New API Key</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createApiKey">
                <div class="mb-3">
                    <CFormLabel for="name">Name</CFormLabel>
                    <CFormInput id="name" v-model="name" placeholder="My API Key" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="description">Description</CFormLabel>
                    <CFormInput id="description" v-model="description" placeholder="Optional description" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="expire_in">Expires In (days)</CFormLabel>
                    <CFormInput id="expire_in" v-model="expire_in" type="number" min="1" />
                </div>
                <div class="d-flex gap-2">
                    <CButton type="submit" color="primary" :disabled="submitting">
                        <CSpinner v-if="submitting" size="sm" class="me-1" />Create
                    </CButton>
                    <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
                </div>
            </CForm>
            <CAlert v-if="result.status === 201" color="success" class="mt-3">
                API key created. Save it now — it will not be shown again.
                <div class="mt-2">
                    <code class="user-select-all d-block" style="word-break: break-all;">{{ result.json.api_token
                        }}</code>
                    <CButton size="sm" color="success" variant="outline" class="mt-2" @click="copyToken">
                        {{ copied ? 'Copied!' : 'Copy' }}
                    </CButton>
                </div>
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

const name = ref('')
const description = ref('')
const expire_in = ref(90)
const result = ref({})
const copied = ref(false)
const submitting = ref(false)

const copyToken = async () => {
    await navigator.clipboard.writeText(result.value.json.api_token)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
}

const createApiKey = async () => {
    copied.value = false
    submitting.value = true
    const res = await apiFetch('/api/v1/users/apitokens', {
        method: 'POST',
        body: JSON.stringify({
            name: name.value,
            description: description.value,
            expire_in: parseFloat(expire_in.value),
        }),
    })
    result.value = { status: res.status, json: await res.json() }
    submitting.value = false
}
</script>
