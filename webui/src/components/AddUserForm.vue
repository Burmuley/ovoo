<template>
    <CCard style="max-width: 540px;">
        <CCardHeader class="fw-semibold">Add New User</CCardHeader>
        <CCardBody>
            <CForm @submit.prevent="createUser">
                <div class="mb-3">
                    <CFormLabel>Type</CFormLabel>
                    <Dropdown text="Select type" :items="userTypes" @filter-selected="userTypeSelected = $event" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="login">Login</CFormLabel>
                    <CFormInput id="login" v-model="login" placeholder="username" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="first_name">First Name</CFormLabel>
                    <CFormInput id="first_name" v-model="first_name" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="last_name">Last Name</CFormLabel>
                    <CFormInput id="last_name" v-model="last_name" />
                </div>
                <div class="mb-3">
                    <CFormLabel for="password">Password</CFormLabel>
                    <CFormInput id="password" v-model="password" type="password" />
                </div>
                <div class="d-flex gap-2">
                    <CButton type="submit" color="primary" :disabled="submitting">
                        <CSpinner v-if="submitting" size="sm" class="me-1" />Create
                    </CButton>
                    <CButton color="secondary" variant="outline" @click="emit('done')">Cancel</CButton>
                </div>
            </CForm>
            <CAlert v-if="result.status === 201" color="success" class="mt-3">
                User <strong>{{ result.json.login }}</strong> was successfully created.
            </CAlert>
            <CAlert v-else-if="result.status" color="danger" class="mt-3">
                <div v-for="error in result.json.errors" :key="error.detail">{{ error.detail }}</div>
            </CAlert>
        </CCardBody>
    </CCard>
</template>

<script setup>
import { ref } from 'vue'
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['done'])

const userTypes = [
    { id: 'regular', text: 'regular' },
    { id: 'admin', text: 'admin' },
    { id: 'milter', text: 'milter' },
]
const userTypeSelected = ref('')
const login = ref('')
const first_name = ref('')
const last_name = ref('')
const password = ref('')
const result = ref({})
const submitting = ref(false)

const createUser = async () => {
    submitting.value = true
    const res = await apiFetch('/api/v1/users', {
        method: 'POST',
        body: JSON.stringify({
            login: login.value,
            first_name: first_name.value,
            last_name: last_name.value,
            type: userTypeSelected.value,
            password: password.value,
        }),
    })
    result.value = { status: res.status, json: await res.json() }
    submitting.value = false
}
</script>
