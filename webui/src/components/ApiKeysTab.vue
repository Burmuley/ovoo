<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">API Keys</span>
            <CButton color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Name</CTableHeaderCell>
                        <CTableHeaderCell>Description</CTableHeaderCell>
                        <CTableHeaderCell>Expires</CTableHeaderCell>
                        <CTableHeaderCell>Status</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="key in apiKeys" :key="key.id">

                        <template v-if="editingId === key.id">
                            <CTableDataCell>
                                <CFormInput v-model="editForm.name" size="sm" placeholder="Name" @keyup.enter="saveEdit(key.id)" />
                            </CTableDataCell>
                            <CTableDataCell>
                                <CFormInput v-model="editForm.description" size="sm" placeholder="Description" @keyup.enter="saveEdit(key.id)" />
                            </CTableDataCell>
                            <CTableDataCell>{{ moment(key.expiration).format('LLL') }}</CTableDataCell>
                            <CTableDataCell>
                                <CBadge :color="key.active ? 'success' : 'danger'">
                                    {{ key.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton
                                    color="success"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    :disabled="saving"
                                    @click="saveEdit(key.id)"
                                >
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton
                                    color="secondary"
                                    size="sm"
                                    variant="outline"
                                    @click="cancelEdit"
                                >
                                    <CIcon icon="cilX" />
                                </CButton>
                            </CTableDataCell>
                        </template>

                        <template v-else>
                            <CTableDataCell>{{ key.name }}</CTableDataCell>
                            <CTableDataCell>{{ key.description }}</CTableDataCell>
                            <CTableDataCell>{{ moment(key.expiration).format('LLL') }}</CTableDataCell>
                            <CTableDataCell>
                                <CBadge :color="key.active ? 'success' : 'danger'">
                                    {{ key.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <template v-if="key.active">
                                    <CButton
                                        color="primary"
                                        size="sm"
                                        variant="outline"
                                        class="me-1"
                                        @click="startEdit(key)"
                                    >
                                        <CIcon icon="cilPencil" />
                                    </CButton>
                                    <CButton
                                        color="warning"
                                        size="sm"
                                        variant="outline"
                                        class="me-1"
                                        @click="confirmingId = key.id"
                                    >
                                        <CIcon icon="cilBan" />
                                    </CButton>
                                </template>
                                <CButton
                                    color="danger"
                                    size="sm"
                                    variant="outline"
                                    @click="deleteApiKey(key.id)"
                                >
                                    <CIcon icon="cilTrash" />
                                </CButton>
                            </CTableDataCell>
                        </template>

                    </CTableRow>
                </CTableBody>
            </CTable>
        </CCardBody>
    </CCard>

    <CModal :visible="confirmingId !== null" @close="confirmingId = null">
        <CModalHeader>
            <CModalTitle>Deactivate API Key</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Are you sure you want to deactivate this API key? It will no longer be usable for authentication.
        </CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingId = null">Cancel</CButton>
            <CButton color="warning" :disabled="saving" @click="deactivate(confirmingId)">Yes, deactivate</CButton>
        </CModalFooter>
    </CModal>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import moment from 'moment'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-clicked'])
const apiKeys = ref([])
const editingId = ref(null)
const editForm = ref({ name: '', description: '' })
const confirmingId = ref(null)
const saving = ref(false)

const load = async () => {
    const res = await apiFetch('/api/v1/users/apitokens')
    apiKeys.value = await res.json()
}

const startEdit = (key) => {
    editingId.value = key.id
    editForm.value = { name: key.name, description: key.description ?? '' }
}

const cancelEdit = () => {
    editingId.value = null
}

const saveEdit = async (id) => {
    saving.value = true
    await apiFetch(`/api/v1/users/apitokens/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: editForm.value.name, description: editForm.value.description }),
    })
    saving.value = false
    editingId.value = null
    await load()
}

const deactivate = async (id) => {
    saving.value = true
    await apiFetch(`/api/v1/users/apitokens/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active: false }),
    })
    saving.value = false
    confirmingId.value = null
    await load()
}

const deleteApiKey = async (id) => {
    await apiFetch(`/api/v1/users/apitokens/${id}`, { method: 'DELETE' })
    await load()
}

onMounted(load)
</script>
