<template>
<CCard>
    <CCardHeader class="d-flex align-items-center justify-content-between">
        <div class="d-flex align-items-center">
            <span class="fw-semibold">Protected Addresses</span>
            <InfoPopover
                description="A protected address is your real email inbox that you want to keep private. Aliases forward mail here so senders only ever see the alias, never your actual address. Deactivating a protected address stops delivery for all aliases that point to it." />
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
                    <CTableHeaderCell>Email</CTableHeaderCell>
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
                <EmptyState v-else-if="praddrs.length === 0" icon="cilShieldAlt"
                    message="No protected addresses yet. Add one to start creating aliases." action-label="Add Address"
                    :colspan="3" @action-clicked="showAddModal = true" />
                <template v-else>
                    <CTableRow v-for="addr in praddrs" :key="addr.id">
                        <CTableDataCell>
                            <span v-c-tooltip="addr.email" class="d-inline-block text-truncate"
                                style="max-width:260px;">{{ addr.email }}</span>
                            <div class="text-body-secondary" style="font-size: 0.75rem;">{{
                                addr.metadata?.comment || ' ' }}</div>
                        </CTableDataCell>
                        <CTableDataCell class="text-center text-nowrap">
                            <CBadge :color="addr.active ? 'success' : 'danger'">
                                {{ addr.active ? 'Active' : 'Inactive' }}
                            </CBadge>
                        </CTableDataCell>
                        <CTableDataCell class="text-end text-nowrap">
                            <CButton v-c-tooltip="'Edit'" color="primary" size="sm" variant="outline" class="me-1"
                                @click="startEdit(addr)">
                                <CIcon icon="cilPencil" />
                            </CButton>
                            <CButton v-if="addr.active" v-c-tooltip="'Deactivate'" color="warning" size="sm"
                                variant="outline" class="me-1" @click="confirmingDeactivateId = addr.id">
                                <CIcon icon="cilBan" />
                            </CButton>
                            <CButton v-else v-c-tooltip="'Activate'" color="success" size="sm" variant="outline"
                                class="me-1" @click="confirmingActivateId = addr.id">
                                <CIcon icon="cilCheckCircle" />
                            </CButton>
                            <CButton v-c-tooltip="'Delete'" color="danger" size="sm" variant="outline"
                                @click="deletingId = addr.id">
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

<CModal :visible="editingAddr !== null" @close="editingAddr = null">
    <CModalHeader>
        <CModalTitle>Edit Protected Address</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <CFormLabel>Comment</CFormLabel>
        <CFormInput v-model="editComment" placeholder="Comment (optional)" />
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="editingAddr = null">Cancel</CButton>
        <CButton color="primary" :disabled="saving" @click="saveEdit">Save</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="deletingId !== null" @close="deletingId = null">
    <CModalHeader>
        <CModalTitle>Delete Protected Address</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Delete <strong>{{ deletingAddr?.email }}</strong>? This action cannot be undone.
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
        <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
    </CModalFooter>
</CModal>

<CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
    <CModalHeader>
        <CModalTitle>Deactivate Protected Address</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Deactivate <strong>{{ confirmingDeactivateAddr?.email }}</strong>? Aliases forwarding to it will stop
        delivering
        email.
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
        <CModalTitle>Activate Protected Address</CModalTitle>
    </CModalHeader>
    <CModalBody>
        Activate <strong>{{ confirmingActivateAddr?.email }}</strong>?
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
        <CModalTitle>Add New Protected Address</CModalTitle>
    </CModalHeader>
    <CModalBody>
        <AddPrAddrForm v-if="showAddModal" @created="onAddrCreated" @done="showAddModal = false" />
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
import AddPrAddrForm from './AddPrAddrForm.vue'

const { showToast } = useToast()

const praddrs = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const searchQuery = ref('')
const loading = ref(true)
const showAddModal = ref(false)
const editingAddr = ref(null)
const editComment = ref('')
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)
const apiError = ref(null)

const deletingAddr = computed(() => praddrs.value.find(a => a.id === deletingId.value))
const confirmingDeactivateAddr = computed(() => praddrs.value.find(a => a.id === confirmingDeactivateId.value))
const confirmingActivateAddr = computed(() => praddrs.value.find(a => a.id === confirmingActivateId.value))

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

let searchDebounce = null

const load = async () => {
    loading.value = true
    const params = new URLSearchParams({ page: currentPage.value })
    if (searchQuery.value) params.set('q', searchQuery.value)
    const res = await apiFetch('/api/v1/praddrs?' + params)
    const data = await res.json()
    praddrs.value = data.protected_addresses
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

function onAddrCreated(email) {
    showAddModal.value = false
    showToast(`Protected address ${email} created.`)
    load()
}

const startEdit = (addr) => {
    editingAddr.value = addr
    editComment.value = addr.metadata?.comment ?? ''
}

const saveEdit = async () => {
    saving.value = true
    const res = await apiFetch(`/api/v1/praddrs/${editingAddr.value.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ metadata: { comment: editComment.value } }),
    })
    saving.value = false
    if (!res.ok) { await handleApiError(res); return }
    editingAddr.value = null
    showToast('Address updated.')
    await load()
}

const setActive = async (id, active) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/praddrs/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active }),
    })
    saving.value = false
    confirmingDeactivateId.value = null
    confirmingActivateId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast(active ? 'Address activated.' : 'Address deactivated.')
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/praddrs/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('Address deleted.')
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
