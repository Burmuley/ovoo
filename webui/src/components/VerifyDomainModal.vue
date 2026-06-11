<template>
<CModal :visible="domain !== null" @close="emit('close')">
    <CModalHeader>
        <CModalTitle>Verify Domain — {{ domain?.name }}</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <template v-if="domain?.verification_data">
            <p>Make sure the following DNS record is published, then click <strong>Verify</strong>.</p>
            <div class="p-3 rounded mb-3"
                style="background: var(--cui-tertiary-bg, #f8f9fa); font-family: monospace; font-size: 0.875rem;">
                <div class="mb-2">
                    <span class="text-body-secondary">Record Type:</span>
                    {{ domain.verification_data.record_type.toUpperCase() }}
                </div>
                <div class="mb-2 d-flex align-items-start justify-content-between gap-2">
                    <div style="min-width: 0; word-break: break-all;"><span class="text-body-secondary">Host /
                            Name:</span> {{ domain.verification_data.name }}</div>
                    <CButton size="sm" color="success" variant="outline" class="flex-shrink-0" @click="copyName">
                        {{ copiedName ? 'Copied!' : 'Copy' }}
                    </CButton>
                </div>
                <div class="d-flex align-items-start justify-content-between gap-2">
                    <div style="min-width: 0; word-break: break-all;"><span class="text-body-secondary">Value:</span> {{
                        domain.verification_data.value }}</div>
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
                            {{ sysInfo.dkim_selector }}._domainkey.{{ domain.name }}
                        </div>
                        <CButton size="sm" color="success" variant="outline" class="flex-shrink-0"
                            @click="copyDkimName">
                            {{ copiedDkimName ? 'Copied!' : 'Copy' }}
                        </CButton>
                    </div>
                    <div class="d-flex align-items-start justify-content-between gap-2">
                        <div style="min-width: 0; word-break: break-all;">
                            <span class="text-body-secondary">Value:</span>
                            {{ sysInfo.dkim_selector }}._domainkey.{{ sysInfo.dkim_domain }}
                        </div>
                        <CButton size="sm" color="success" variant="outline" class="flex-shrink-0"
                            @click="copyDkimValue">
                            {{ copiedDkimValue ? 'Copied!' : 'Copy' }}
                        </CButton>
                    </div>
                </div>
            </template>
        </template>
        <CAlert v-if="verifyResult?.verified" color="success">
            Domain verified successfully.
        </CAlert>
        <CAlert v-else-if="verifyResult" color="warning" style="word-break: break-word;">
            {{ verifyResult.verification_data?.last_verification_result ?? verifyFailMsg }}
        </CAlert>
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="emit('close')">Close</CButton>
        <CButton v-if="!verifyResult?.verified" color="primary" :disabled="verifying" @click="performVerify">
            <CSpinner v-if="verifying" size="sm" class="me-1" />Verify
        </CButton>
    </CModalFooter>
</CModal>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'

const props = defineProps({ domain: { type: Object, default: null } })
const emit = defineEmits(['close', 'verify-complete'])
const { showToast } = useToast()

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

watch(() => props.domain, (val) => {
    if (val !== null) {
        verifyResult.value = null
        copiedName.value = false
        copiedValue.value = false
    }
})

const copyName = async () => {
    await navigator.clipboard.writeText(props.domain.verification_data.name)
    copiedName.value = true
    setTimeout(() => { copiedName.value = false }, 2000)
}
const copyValue = async () => {
    await navigator.clipboard.writeText(props.domain.verification_data.value)
    copiedValue.value = true
    setTimeout(() => { copiedValue.value = false }, 2000)
}
const copyDkimName = async () => {
    await navigator.clipboard.writeText(
        `${sysInfo.value.dkim_selector}._domainkey.${props.domain.name}`
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

const performVerify = async () => {
    verifying.value = true
    const res = await apiFetch(`/api/v1/domains/${props.domain.id}/verify`, { method: 'POST' })
    const data = await res.json()
    verifying.value = false
    verifyResult.value = data
    if (res.ok && data.verified) showToast('Domain verified successfully.')
    emit('verify-complete')
}
</script>
