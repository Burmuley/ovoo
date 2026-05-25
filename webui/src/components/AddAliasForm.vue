<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New Alias</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createAlias">
                <div class="mb-3">
                    <CFormLabel>Protected Address</CFormLabel>
                    <Dropdown text="Select address" :items="praddrs" @filter-selected="praddrSelected = $event" />
                </div>
                <div class="mb-3" v-if="domains.length > 1">
                    <CFormLabel>Domain</CFormLabel>
                    <Dropdown text="Select domain" :items="domainItems" @filter-selected="domainSelected = $event" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="svcname">Service Name</CFormLabel>
                    <CFormInput id="svcname" v-model="svcname" placeholder="e.g. GitHub" />
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
                Alias was successfully created.
                <div class="mt-2 d-flex align-items-center gap-2">
                    <code class="user-select-all">{{ result.json.email }}</code>
                    <CButton size="sm" color="success" variant="outline" @click="copyAlias">
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
import { ref, onMounted } from 'vue'
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['done'])

const praddrs = ref([])
const praddrSelected = ref('')
const domains = ref([])
const domainItems = ref([])
const domainSelected = ref('')
const svcname = ref('')
const comment = ref('')
const result = ref({})
const copied = ref(false)

const load = async () => {
    const res = await apiFetch('/api/v1/praddrs')
    const data = await res.json()
    praddrs.value = data.protected_addresses.map(a => ({ id: a.id, text: a.email }))
}

const loadDomains = async () => {
    const res = await apiFetch('/api/v1/domains')
    const data = await res.json()
    domains.value = data.domains
    domainItems.value = data.domains.map(d => ({ id: d, text: d }))
    if (domains.value.length > 0) {
        domainSelected.value = domains.value[0]
    }
}

const copyAlias = async () => {
    await navigator.clipboard.writeText(result.value.json.email)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
}

const createAlias = async () => {
    copied.value = false
    const body = {
        protected_address_id: praddrSelected.value.toString(),
        metadata: {
            service_name: svcname.value,
            comment: comment.value,
        },
    }
    if (domainSelected.value) {
        body.domain = domainSelected.value
    }
    const res = await apiFetch('/api/v1/aliases', {
        method: 'POST',
        body: JSON.stringify(body),
    })
    result.value = { status: res.status, json: await res.json() }
}

onMounted(() => {
    load()
    loadDomains()
})
</script>
