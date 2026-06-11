<template>
<div>
    <div class="d-flex justify-content-between align-items-center mb-3">
        <span class="fw-semibold">{{ currentStep === 1 ? 'Add New Domain' : 'Verify Your Domain' }}</span>
        <small class="text-body-secondary">Step {{ currentStep }} of 2</small>
    </div>

    <!-- Step 1: General Settings -->
    <CForm v-if="currentStep === 1" @submit.prevent="createDomain">
        <div class="mb-3">
            <CFormLabel for="domainName">Domain Name</CFormLabel>
            <CFormInput id="domainName" v-model="name" placeholder="e.g. example.com" required />
        </div>
        <div class="mb-3" v-if="props.userInfo.type === 'admin'">
            <CFormLabel for="domainType">Type</CFormLabel>
            <CFormSelect id="domainType" v-model="domainType">
                <option value="personal">Personal</option>
                <option value="global">Global</option>
            </CFormSelect>
        </div>
        <div class="mb-3">
            <CFormLabel>Verification Method</CFormLabel>
            <div class="d-flex flex-column gap-2">
                <CFormCheck type="radio" id="verifyTxt" name="verificationType" value="dns_txt"
                    v-model="verificationType" label="DNS TXT record" />
                <CFormCheck type="radio" id="verifyCname" name="verificationType" value="dns_cname"
                    v-model="verificationType" label="DNS CNAME record" />
            </div>
        </div>
        <CAlert v-if="errorMessages.length" color="danger" class="mb-3">
            <div v-for="msg in errorMessages" :key="msg">{{ msg }}</div>
        </CAlert>
        <div class="d-flex gap-2">
            <CButton type="submit" color="primary" :disabled="submitting">
                <CSpinner v-if="submitting" size="sm" class="me-1" />Next
            </CButton>
            <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
        </div>
    </CForm>

    <!-- Step 2: DNS Verification Instructions + Inline Verify -->
    <div v-else-if="currentStep === 2">
        <p>Create the DNS record below to verify ownership of <strong>{{ createdDomain.name }}</strong>.</p>
        <div class="p-3 rounded mb-3"
            style="background: var(--cui-tertiary-bg, #f8f9fa); font-family: monospace; font-size: 0.875rem;">
            <div class="mb-2">
                <span class="text-body-secondary">Record Type:</span>
                {{ createdDomain.verification_data.record_type.toUpperCase() }}
            </div>
            <div class="mb-2 d-flex align-items-start justify-content-between gap-2">
                <div style="min-width: 0; word-break: break-all;"><span class="text-body-secondary">Host /
                        Name:</span> {{ createdDomain.verification_data.name }}</div>
                <CButton size="sm" color="success" variant="outline" class="flex-shrink-0" @click="copyName">
                    {{ copiedName ? 'Copied!' : 'Copy' }}
                </CButton>
            </div>
            <div class="d-flex align-items-start justify-content-between gap-2">
                <div style="min-width: 0; word-break: break-all;"><span class="text-body-secondary">Value:</span> {{
                    createdDomain.verification_data.value }}
                </div>
                <CButton size="sm" color="success" variant="outline" class="flex-shrink-0" @click="copyValue">
                    {{ copiedValue ? 'Copied!' : 'Copy' }}
                </CButton>
            </div>
        </div>
        <template v-if="sysInfo">
            <p>Also add this CNAME record to enable mail delivery for this domain from Ovoo hosts:</p>
            <div class="p-3 rounded mb-3"
                style="background: var(--cui-tertiary-bg, #f8f9fa); font-family: monospace; font-size: 0.875rem;">
                <div class="mb-2">
                    <span class="text-body-secondary">Record Type:</span> CNAME
                </div>
                <div class="mb-2 d-flex align-items-start justify-content-between gap-2">
                    <div style="min-width: 0; word-break: break-all;">
                        <span class="text-body-secondary">Host / Name:</span>
                        {{ sysInfo.dkim_selector }}._domainkey.{{ createdDomain.name }}
                    </div>
                    <CButton size="sm" color="success" variant="outline" class="flex-shrink-0" @click="copyDkimName">
                        {{ copiedDkimName ? 'Copied!' : 'Copy' }}
                    </CButton>
                </div>
                <div class="d-flex align-items-start justify-content-between gap-2">
                    <div style="min-width: 0; word-break: break-all;">
                        <span class="text-body-secondary">Value:</span>
                        {{ sysInfo.dkim_selector }}._domainkey.{{ sysInfo.dkim_domain }}
                    </div>
                    <CButton size="sm" color="success" variant="outline" class="flex-shrink-0" @click="copyDkimValue">
                        {{ copiedDkimValue ? 'Copied!' : 'Copy' }}
                    </CButton>
                </div>
            </div>
        </template>
        <p class="text-body-secondary small">
            You can verify now or close this and verify later using the
            <strong>Verify</strong> button in the Domains list.
        </p>
        <div class="d-flex gap-2">
            <CButton v-if="!verifyResult?.verified" color="primary" :disabled="verifying" @click="verifyDomain">
                <CSpinner v-if="verifying" size="sm" class="me-1" />Verify
            </CButton>
            <CButton color="secondary" variant="outline" @click="emit('done')">Close</CButton>
        </div>
        <CAlert v-if="verifyResult?.verified" color="success" class="mt-3">
            Domain verified successfully.
        </CAlert>
        <CAlert v-else-if="verifyResult" color="warning" class="mt-3" style="word-break: break-word;">
            {{ verifyResult.verification_data?.last_verification_result ?? verifyFailMsg }}
        </CAlert>
    </div>
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['done'])
const { showToast } = useToast()

const currentStep = ref(1)
const name = ref('')
const domainType = ref('personal')
const verificationType = ref('dns_txt')
const submitting = ref(false)
const errorMessages = ref([])
const createdDomain = ref(null)
const verifying = ref(false)
const verifyResult = ref(null)
const copiedName = ref(false)
const copiedValue = ref(false)
const copiedDkimName = ref(false)
const copiedDkimValue = ref(false)
const sysInfo = ref(null)
const verifyFailMsg = 'Verification failed. Check that the DNS record is published and try again.'

onMounted(async () => {
    const res = await apiFetch('/api/v1/sysinfo')
    if (res?.ok) sysInfo.value = await res.json()
})

const copyName = async () => {
    await navigator.clipboard.writeText(createdDomain.value.verification_data.name)
    copiedName.value = true
    setTimeout(() => { copiedName.value = false }, 2000)
}
const copyValue = async () => {
    await navigator.clipboard.writeText(createdDomain.value.verification_data.value)
    copiedValue.value = true
    setTimeout(() => { copiedValue.value = false }, 2000)
}
const copyDkimName = async () => {
    await navigator.clipboard.writeText(
        `${sysInfo.value.dkim_selector}._domainkey.${createdDomain.value.name}`
    )
    copiedDkimName.value = true
    setTimeout(() => { copiedDkimName.value = false }, 2000)
}
const copyDkimValue = async () => {
    await navigator.clipboard.writeText(
        `${sysInfo.value.dkim_selector}._domainkey.${sysInfo.value.dkim_domain}`
    )
    copiedDkimValue.value = true
    setTimeout(() => { copiedDkimValue.value = false }, 2000)
}

const createDomain = async () => {
    submitting.value = true
    errorMessages.value = []
    const body = { name: name.value, type: domainType.value }
    body.verification_type = verificationType.value
    const res = await apiFetch('/api/v1/domains', {
        method: 'POST',
        body: JSON.stringify(body),
    })
    const data = await res.json()
    submitting.value = false
    if (!res.ok) {
        errorMessages.value = data.errors?.map(e => e.detail) ?? ['An unexpected error occurred.']
        return
    }
    createdDomain.value = data
    currentStep.value = 2
}

const verifyDomain = async () => {
    verifying.value = true
    const res = await apiFetch(`/api/v1/domains/${createdDomain.value.id}/verify`, { method: 'POST' })
    const data = await res.json()
    verifying.value = false
    verifyResult.value = data
    if (res.ok && data.verified) {
        showToast('Domain verified successfully.')
    }
}
</script>
