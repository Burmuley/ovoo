<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New Alias</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createAlias">
                <div class="mb-3">
                    <CFormLabel>Protected Address</CFormLabel>
                    <Dropdown text="Select address" :items="praddrs" @filter-selected="praddrSelected = $event" />
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
                Alias <strong>{{ result.json.email }}</strong> was successfully created.
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
const svcname = ref('')
const comment = ref('')
const result = ref({})

const load = async () => {
    const res = await apiFetch('/api/v1/praddrs')
    const data = await res.json()
    praddrs.value = data.protected_addresses.map(a => ({ id: a.id, text: a.email }))
}

const createAlias = async () => {
    const res = await apiFetch('/api/v1/aliases', {
        method: 'POST',
        body: JSON.stringify({
            protected_address_id: praddrSelected.value.toString(),
            metadata: {
                service_name: svcname.value,
                comment: comment.value,
            },
        }),
    })
    result.value = { status: res.status, json: await res.json() }
}

onMounted(load)
</script>
