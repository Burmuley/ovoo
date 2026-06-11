<template>
<CCard>
    <CCardHeader class="d-flex align-items-center justify-content-between">
        <div class="d-flex align-items-center">
            <span class="fw-semibold">API Keys</span>
            <InfoPopover
                description="API keys let you authenticate with the Ovoo REST API from scripts or external applications using Bearer token authentication. Each key has an expiration date and can be deactivated or deleted at any time. Please note for security purposes newly created API key values can only be visible right after it was created, this value is not stored in the database and can not be retrieved later." />
        </div>
        <CButton color="primary" size="sm" @click="showAddModal = true">
            <CIcon icon="cilPlus" /> Add
        </CButton>
    </CCardHeader>
    <CCardBody class="p-0">
        <CTable hover responsive class="mb-0">
            <CTableHead>
                <CTableRow>
                    <CTableHeaderCell>Name</CTableHeaderCell>
                    <CTableHeaderCell class="text-center" style="width: 1%; white-space: nowrap;">Status
                    </CTableHeaderCell>
                    <CTableHeaderCell style="width: 1%; white-space: nowrap;"></CTableHeaderCell>
                </CTableRow>
            </CTableHead>
            <CTableBody>
                <template v-if="loading">
                    <CTableRow v-for="n in 3" :key="n">
                        <CTableDataCell v-for="c in 3" :key="c">
                            <div class="placeholder-glow"><span class="placeholder col-8"></span></div>
                        </CTableDataCell>
                    </CTableRow>
                </template>
                <EmptyState v-else-if="apiKeys.length === 0" icon="cilCode"
                    message="No API keys yet. Create one to access the API programmatically." action-label="Add API Key"
                    :colspan="3" @action-clicked="showAddModal = true" />
                <template v-else>
                    <CTableRow v-for="key in apiKeys" :key="key.id">
                        <CTableDataCell>
                            <div>{{ key.name }}</div>
                            <div class="text-body-secondary" style="font-size: 0.75rem;">{{
                                key.description || ' ' }}</div>
                            <div class="text-body-secondary" style="font-size: 0.75rem;">Expires: {{
                                moment(key.expiration).format('LLL') }}</div>
                        </CTableDataCell>
                        <CTableDataCell class="text-center text-nowrap">
                            <CBadge :color="key.active ? 'success' : 'danger'">
                                {{ key.active ? 'Active' : 'Inactive' }}
                            </CBadge>
                        </CTableDataCell>
                        <CTableDataCell class="text-end text-nowrap">
                            <template v-if="key.active">
                                <CButton color="primary" size="sm" variant="outline" class="me-1"
                                    @click="startEdit(key)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                                <CButton color="warning" size="sm" variant="outline" class="me-1"
                                    @click="confirmingId = key.id">
                                    <CIcon icon="cilBan" />
                                </CButton>
                            </template>
                            <CButton color="danger" size="sm" variant="outline" @click="deletingId = key.id">
                                <CIcon icon="cilTrash" />
                            </CButton>
                        </CTableDataCell>
                    </CTableRow>
                </template>
            </CTableBody>
        </CTable>
    </CCardBody>
</CCard>

<CModal :visible="editingKey !== null" @close="editingKey = null">
    <CModalHeader>
        <CModalTitle>Edit API Key</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <CFormLabel>Name</CFormLabel>
        <CFormInput v-model="editForm.name" placeholder="Name" class="mb-3" />
        <CFormLabel>Description</CFormLabel>
        <CFormInput v-model="editForm.description" placeholder="Description (optional)" />
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="editingKey = null">Cancel</CButton>
        <CButton color="primary" :disabled="saving" @click="saveEdit">Save</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="deletingId !== null" @close="deletingId = null">
    <CModalHeader>
        <CModalTitle>Delete API Key</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Delete API key <strong>{{ deletingKey?.name }}</strong>? This action cannot be undone.
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
        <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="confirmingId !== null" @close="confirmingId = null">
    <CModalHeader>
        <CModalTitle>Deactivate API Key</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Deactivate <strong>{{ confirmingKey?.name }}</strong>? It will no longer be usable for authentication.
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="confirmingId = null">Cancel</CButton>
        <CButton color="warning" :disabled="saving" @click="deactivate(confirmingId)">Yes, deactivate</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="apiError !== null" @close="apiError = null">
    <CModalHeader>
        <CModalTitle>Error</CModalTitle>
    </CModalHeader>
    <CModalBody>{{ apiError }}</CModalBody>
    <CModalFooter>
        <CButton color="secondary" @click="apiError = null">Close</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="showAddModal" @close="onFormDone" size="lg">
    <CModalHeader>
        <CModalTitle>Add New API Key</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <AddApiKeyForm v-if="showAddModal" @done="onFormDone" />
    </CModalBody>
</CModal>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import moment from 'moment'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'
import EmptyState from './EmptyState.vue'
import InfoPopover from './InfoPopover.vue'
import AddApiKeyForm from './AddApiKeyForm.vue'

const { showToast } = useToast()

const apiKeys = ref([])
const loading = ref(true)
const showAddModal = ref(false)
const editingKey = ref(null)
const editForm = ref({ name: '', description: '' })
const confirmingId = ref(null)
const deletingId = ref(null)
const saving = ref(false)
const apiError = ref(null)

const deletingKey = computed(() => apiKeys.value.find(k => k.id === deletingId.value))
const confirmingKey = computed(() => apiKeys.value.find(k => k.id === confirmingId.value))

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

const load = async () => {
    loading.value = true
    const res = await apiFetch('/api/v1/users/apitokens')
    apiKeys.value = await res.json()
    loading.value = false
}

function onFormDone() {
    showAddModal.value = false
    load()
}

const startEdit = (key) => {
    editingKey.value = key
    editForm.value = { name: key.name, description: key.description ?? '' }
}

const saveEdit = async () => {
    saving.value = true
    const res = await apiFetch(`/api/v1/users/apitokens/${editingKey.value.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ name: editForm.value.name, description: editForm.value.description }),
    })
    saving.value = false
    if (!res.ok) { await handleApiError(res); return }
    editingKey.value = null
    showToast('API key updated.')
    await load()
}

const deactivate = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/users/apitokens/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active: false }),
    })
    saving.value = false
    confirmingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('API key deactivated.')
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/users/apitokens/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('API key deleted.')
    await load()
}

function onDeleteKey(e) {
    if (e.key === 'Enter') { e.preventDefault(); performDelete(deletingId.value) }
}
watch(deletingId, id => {
    if (id !== null) document.addEventListener('keydown', onDeleteKey)
    else document.removeEventListener('keydown', onDeleteKey)
})
onUnmounted(() => document.removeEventListener('keydown', onDeleteKey))

onMounted(load)
</script>
