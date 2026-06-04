<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New Domain</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createDomain">
                <div class="mb-3">
                    <CFormLabel for="domainName">Domain Name</CFormLabel>
                    <CFormInput id="domainName" v-model="name" placeholder="e.g. example.com" />
                </div>
                <div class="mb-3" v-if="props.userInfo.type === 'admin'">
                    <CFormLabel for="domainType">Type</CFormLabel>
                    <CFormSelect id="domainType" v-model="domainType">
                        <option value="personal">Personal</option>
                        <option value="global">Global</option>
                    </CFormSelect>
                </div>
                <div class="d-flex gap-2">
                    <CButton type="submit" color="primary" :disabled="submitting">
                        <CSpinner v-if="submitting" size="sm" class="me-1" />Create
                    </CButton>
                    <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
                </div>
            </CForm>
            <CAlert v-if="result.status === 201" color="success" class="mt-3">
                Domain <strong>{{ result.json.name }}</strong> was successfully created.
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

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['done'])

const name = ref('')
const domainType = ref('personal')
const result = ref({})
const submitting = ref(false)

const createDomain = async () => {
    submitting.value = true
    const body = { name: name.value, type: domainType.value }
    const res = await apiFetch('/api/v1/domains', {
        method: 'POST',
        body: JSON.stringify(body),
    })
    result.value = { status: res.status, json: await res.json() }
    submitting.value = false
}
</script>
