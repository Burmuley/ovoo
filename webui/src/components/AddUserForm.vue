<template>
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
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['done', 'created'])

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
const errorMessage = ref('')
const submitting = ref(false)

const createUser = async () => {
    errorMessage.value = ''
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
    const json = await res.json()
    submitting.value = false
    if (res.status === 201) {
        emit('created', json.login)
    } else {
        errorMessage.value = json.errors?.[0]?.detail ?? 'An unexpected error occurred'
    }
}
</script>
