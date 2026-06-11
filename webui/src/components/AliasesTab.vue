<template>
<CCard>
    <CCardHeader class="d-flex align-items-center justify-content-between">
        <div class="d-flex align-items-center">
            <span class="fw-semibold">Aliases</span>
            <InfoPopover
                description="An alias is a randomly generated email address that forwards incoming messages to your protected (real) address. Share the alias instead of your real email — if it gets spammed, deactivate or delete it without touching your real inbox." />
        </div>
        <div class="d-flex align-items-center gap-2">
            <div class="input-group input-group-sm" style="width: 220px;">
                <span class="input-group-text">
                    <CIcon icon="cilSearch" />
                </span>
                <CFormInput v-model="searchQuery" placeholder="Search…" />
                <CButton v-if="searchQuery" color="secondary" variant="outline" @click="searchQuery = ''">
                    <CIcon icon="cilX" />
                </CButton>
            </div>
            <CButton color="primary" size="sm" @click="showAddModal = true">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </div>
    </CCardHeader>
    <CCardBody class="p-0">
        <CTable hover responsive class="mb-0">
            <CTableHead>
                <CTableRow>
                    <CTableHeaderCell>Alias</CTableHeaderCell>
                    <CTableHeaderCell class="text-center">Forwards To</CTableHeaderCell>
                    <CTableHeaderCell class="text-center" style="width: 1%; white-space: nowrap;">Status
                    </CTableHeaderCell>
                    <CTableHeaderCell style="width: 1%; white-space: nowrap;"></CTableHeaderCell>
                </CTableRow>
            </CTableHead>
            <CTableBody>
                <template v-if="loading">
                    <CTableRow v-for="n in 3" :key="n">
                        <CTableDataCell v-for="c in 4" :key="c">
                            <div class="placeholder-glow"><span class="placeholder col-8"></span></div>
                        </CTableDataCell>
                    </CTableRow>
                </template>
                <EmptyState v-else-if="aliases.length === 0" icon="cilEnvelopeClosed"
                    message="No aliases yet. Create one to get started." action-label="Add Alias" :colspan="4"
                    @action-clicked="showAddModal = true" />
                <template v-else>
                    <CTableRow v-for="alias in aliases" :key="alias.id">
                        <CTableDataCell>
                            <div v-c-tooltip="alias.email" class="text-truncate">{{ alias.email }}</div>
                            <div class="text-body-secondary" style="font-size: 0.75rem;">{{
                                alias.metadata?.service_name ? 'Service: ' + alias.metadata.service_name : ' ' }}</div>
                            <div class="text-body-secondary" style="font-size: 0.75rem;">{{
                                alias.metadata?.comment || ' ' }}</div>
                        </CTableDataCell>
                        <CTableDataCell class="text-center">
                            <span v-c-tooltip="alias.forward_email" class="d-inline-block text-truncate"
                                style="max-width:180px;">{{ alias.forward_email }}</span>
                        </CTableDataCell>
                        <CTableDataCell class="text-center text-nowrap">
                            <CBadge :color="alias.active ? 'success' : 'danger'">
                                {{ alias.active ? 'Active' : 'Inactive' }}
                            </CBadge>
                        </CTableDataCell>
                        <CTableDataCell class="text-end text-nowrap">
                            <CButton v-c-tooltip="'Edit'" color="primary" size="sm" variant="outline" class="me-1"
                                @click="startEdit(alias)">
                                <CIcon icon="cilPencil" />
                            </CButton>
                            <CButton v-if="alias.active" v-c-tooltip="'Deactivate'" color="warning" size="sm"
                                variant="outline" class="me-1" @click="confirmingDeactivateId = alias.id">
                                <CIcon icon="cilBan" />
                            </CButton>
                            <CButton v-else v-c-tooltip="'Activate'" color="success" size="sm" variant="outline"
                                class="me-1" @click="confirmingActivateId = alias.id">
                                <CIcon icon="cilCheckCircle" />
                            </CButton>
                            <CButton v-c-tooltip="'Delete'" color="danger" size="sm" variant="outline"
                                @click="deletingId = alias.id">
                                <CIcon icon="cilTrash" />
                            </CButton>
                        </CTableDataCell>
                    </CTableRow>
                </template>
            </CTableBody>
        </CTable>
    </CCardBody>
    <CCardFooter v-if="paginationMetadata.last_page > 1" class="d-flex justify-content-center">
        <Paginator :current-page="currentPage" :total-pages="paginationMetadata.last_page"
            :total-items="paginationMetadata.total_records" @page-changed="onPageChanged" />
    </CCardFooter>
</CCard>

<CModal :visible="editingAlias !== null" @close="editingAlias = null">
    <CModalHeader>
        <CModalTitle>Edit Alias</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <CFormLabel>Service Name</CFormLabel>
        <CFormInput v-model="editForm.service_name" placeholder="Service name (optional)" class="mb-3" />
        <CFormLabel>Comment</CFormLabel>
        <CFormInput v-model="editForm.comment" placeholder="Comment (optional)" />
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="editingAlias = null">Cancel</CButton>
        <CButton color="primary" :disabled="saving" @click="saveEdit">Save</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="deletingId !== null" @close="deletingId = null">
    <CModalHeader>
        <CModalTitle>Delete Alias</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Delete alias <strong>{{ deletingAlias?.email }}</strong>? This action cannot be undone.
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
        <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
    <CModalHeader>
        <CModalTitle>Deactivate Alias</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Deactivate <strong>{{ confirmingDeactivateAlias?.email }}</strong>? Emails sent to it will stop being
        forwarded.
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="confirmingDeactivateId = null">Cancel</CButton>
        <CButton color="warning" :disabled="saving" @click="setActive(confirmingDeactivateId, false)">Yes,
            deactivate
        </CButton>
    </CModalFooter>
</CModal>

<CModal :visible="confirmingActivateId !== null" @close="confirmingActivateId = null">
    <CModalHeader>
        <CModalTitle>Activate Alias</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Activate <strong>{{ confirmingActivateAlias?.email }}</strong>?
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="confirmingActivateId = null">Cancel</CButton>
        <CButton color="success" :disabled="saving" @click="setActive(confirmingActivateId, true)">Yes, activate
        </CButton>
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

<CModal :visible="showAddModal" @close="showAddModal = false" size="lg">
    <CModalHeader>
        <CModalTitle>Add New Alias</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <AddAliasForm v-if="showAddModal" @created="onAliasCreated" @done="showAddModal = false" />
    </CModalBody>
</CModal>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'
import Paginator from './Paginator.vue'
import EmptyState from './EmptyState.vue'
import InfoPopover from './InfoPopover.vue'
import AddAliasForm from './AddAliasForm.vue'

const { showToast } = useToast()

const aliases = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const searchQuery = ref('')
const loading = ref(true)
const showAddModal = ref(false)
const editingAlias = ref(null)
const editForm = ref({ service_name: '', comment: '' })
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)
const apiError = ref(null)

const deletingAlias = computed(() => aliases.value.find(a => a.id === deletingId.value))
const confirmingDeactivateAlias = computed(() => aliases.value.find(a => a.id === confirmingDeactivateId.value))
const confirmingActivateAlias = computed(() => aliases.value.find(a => a.id === confirmingActivateId.value))

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

let searchDebounce = null

const load = async () => {
    loading.value = true
    const params = new URLSearchParams({ page: currentPage.value })
    if (searchQuery.value) params.set('q', searchQuery.value)
    const res = await apiFetch('/api/v1/aliases?' + params)
    const data = await res.json()
    aliases.value = data.aliases
    paginationMetadata.value = data.pagination_metadata
    loading.value = false
}

watch(searchQuery, () => {
    clearTimeout(searchDebounce)
    searchDebounce = setTimeout(() => {
        currentPage.value = 1
        load()
    }, 300)
})

function onAliasCreated(email) {
    showAddModal.value = false
    showToast(`Alias created: ${email}`)
    load()
}

const startEdit = (alias) => {
    editingAlias.value = alias
    editForm.value = {
        service_name: alias.metadata?.service_name ?? '',
        comment: alias.metadata?.comment ?? '',
    }
}

const saveEdit = async () => {
    saving.value = true
    const res = await apiFetch(`/api/v1/aliases/${editingAlias.value.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
            metadata: {
                service_name: editForm.value.service_name,
                comment: editForm.value.comment,
            },
        }),
    })
    saving.value = false
    if (!res.ok) { await handleApiError(res); return }
    editingAlias.value = null
    showToast('Alias updated.')
    await load()
}

const setActive = async (id, active) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/aliases/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active }),
    })
    saving.value = false
    confirmingDeactivateId.value = null
    confirmingActivateId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast(active ? 'Alias activated.' : 'Alias deactivated.')
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/aliases/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('Alias deleted.')
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

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>
