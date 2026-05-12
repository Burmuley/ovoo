<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Aliases</span>
            <CButton color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Alias</CTableHeaderCell>
                        <CTableHeaderCell>Forwards To</CTableHeaderCell>
                        <CTableHeaderCell>Service</CTableHeaderCell>
                        <CTableHeaderCell>Comment</CTableHeaderCell>
                        <CTableHeaderCell>Status</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="alias in aliases" :key="alias.id">
                        <CTableDataCell>{{ alias.email }}</CTableDataCell>
                        <CTableDataCell>{{ alias.forward_email }}</CTableDataCell>

                        <template v-if="editingId === alias.id">
                            <CTableDataCell>
                                <CFormInput
                                    v-model="editForm.service_name"
                                    size="sm"
                                    placeholder="Service name"
                                    @keyup.enter="saveEdit(alias.id)"
                                />
                            </CTableDataCell>
                            <CTableDataCell>
                                <CFormInput
                                    v-model="editForm.comment"
                                    size="sm"
                                    placeholder="Comment"
                                    @keyup.enter="saveEdit(alias.id)"
                                />
                            </CTableDataCell>
                            <CTableDataCell>
                                <CBadge :color="alias.active ? 'success' : 'danger'">
                                    {{ alias.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton
                                    color="success"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    :disabled="saving"
                                    @click="saveEdit(alias.id)"
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
                            <CTableDataCell>{{ alias.metadata?.service_name }}</CTableDataCell>
                            <CTableDataCell>{{ alias.metadata?.comment }}</CTableDataCell>
                            <CTableDataCell>
                                <CBadge :color="alias.active ? 'success' : 'danger'">
                                    {{ alias.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton
                                    color="primary"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    @click="startEdit(alias)"
                                >
                                    <CIcon icon="cilPencil" />
                                </CButton>
                                <CButton
                                    v-if="alias.active"
                                    color="warning"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    @click="confirmingDeactivateId = alias.id"
                                >
                                    <CIcon icon="cilBan" />
                                </CButton>
                                <CButton
                                    v-else
                                    color="success"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    @click="confirmingActivateId = alias.id"
                                >
                                    <CIcon icon="cilCheckCircle" />
                                </CButton>
                                <CButton
                                    color="danger"
                                    size="sm"
                                    variant="outline"
                                    @click="deletingId = alias.id"
                                >
                                    <CIcon icon="cilTrash" />
                                </CButton>
                            </CTableDataCell>
                        </template>
                    </CTableRow>
                </CTableBody>
            </CTable>
        </CCardBody>
        <CCardFooter v-if="paginationMetadata.last_page > 1" class="d-flex justify-content-center">
            <Paginator
                :current-page="currentPage"
                :total-pages="paginationMetadata.last_page"
                @page-changed="onPageChanged"
            />
        </CCardFooter>
    </CCard>

    <CModal :visible="deletingId !== null" @close="deletingId = null">
        <CModalHeader><CModalTitle>Delete Alias</CModalTitle></CModalHeader>
        <CModalBody>Are you sure you want to delete this alias? This action cannot be undone.</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader><CModalTitle>Deactivate Alias</CModalTitle></CModalHeader>
        <CModalBody>Are you sure you want to deactivate this alias? Emails sent to it will stop being forwarded.</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingDeactivateId = null">Cancel</CButton>
            <CButton color="warning" :disabled="saving" @click="setActive(confirmingDeactivateId, false)">Yes, deactivate</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingActivateId !== null" @close="confirmingActivateId = null">
        <CModalHeader><CModalTitle>Activate Alias</CModalTitle></CModalHeader>
        <CModalBody>Are you sure you want to activate this alias?</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingActivateId = null">Cancel</CButton>
            <CButton color="success" :disabled="saving" @click="setActive(confirmingActivateId, true)">Yes, activate</CButton>
        </CModalFooter>
    </CModal>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-clicked'])
const aliases = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const editingId = ref(null)
const editForm = ref({ service_name: '', comment: '' })
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)

const load = async () => {
    const res = await apiFetch('/api/v1/aliases?page=' + currentPage.value)
    const data = await res.json()
    aliases.value = data.aliases
    paginationMetadata.value = data.pagination_metadata
}

const startEdit = (alias) => {
    editingId.value = alias.id
    editForm.value = {
        service_name: alias.metadata?.service_name ?? '',
        comment: alias.metadata?.comment ?? '',
    }
}

const cancelEdit = () => {
    editingId.value = null
}

const saveEdit = async (id) => {
    saving.value = true
    await apiFetch(`/api/v1/aliases/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({
            metadata: {
                service_name: editForm.value.service_name,
                comment: editForm.value.comment,
            },
        }),
    })
    saving.value = false
    editingId.value = null
    await load()
}

const setActive = async (id, active) => {
    saving.value = true
    await apiFetch(`/api/v1/aliases/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active }),
    })
    saving.value = false
    confirmingDeactivateId.value = null
    confirmingActivateId.value = null
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    await apiFetch(`/api/v1/aliases/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    await load()
}

function onDeleteKey(e) {
    if (e.key === 'Enter') { e.preventDefault(); performDelete(deletingId.value) }
}
watch(deletingId, id => {
    if (id !== null) document.addEventListener('keydown', onDeleteKey)
    else             document.removeEventListener('keydown', onDeleteKey)
})
onUnmounted(() => document.removeEventListener('keydown', onDeleteKey))

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>
