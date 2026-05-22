<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Protected Addresses</span>
            <div class="d-flex align-items-center gap-2">
                <CButton v-if="!showSearch" v-c-tooltip="'Search'" color="secondary" size="sm" variant="ghost"
                    @click="openSearch">
                    <CIcon icon="cilSearch" />
                </CButton>
                <CFormInput v-if="showSearch" ref="searchInputRef" v-model="searchQuery" size="sm" placeholder="Search…"
                    style="width: 180px" />
                <CButton v-if="showSearch" v-c-tooltip="'Clear search'" color="secondary" size="sm" variant="ghost"
                    @click="closeSearch">
                    <CIcon icon="cilX" />
                </CButton>
                <CButton color="primary" size="sm" @click="emit('add-clicked')">
                    <CIcon icon="cilPlus" /> Add
                </CButton>
            </div>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Email</CTableHeaderCell>
                        <CTableHeaderCell>Comment</CTableHeaderCell>
                        <CTableHeaderCell>Status</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="addr in praddrs" :key="addr.id">
                        <CTableDataCell>{{ addr.email }}</CTableDataCell>

                        <template v-if="editingId === addr.id">
                            <CTableDataCell>
                                <CFormInput v-model="editComment" size="sm" placeholder="Comment"
                                    @keyup.enter="saveEdit(addr.id)" />
                            </CTableDataCell>
                            <CTableDataCell>
                                <CBadge :color="addr.active ? 'success' : 'danger'">
                                    {{ addr.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton v-c-tooltip="'Save'" color="success" size="sm" variant="outline" class="me-1"
                                    :disabled="saving" @click="saveEdit(addr.id)">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton v-c-tooltip="'Cancel'" color="secondary" size="sm" variant="outline"
                                    @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </CTableDataCell>
                        </template>

                        <template v-else>
                            <CTableDataCell>{{ addr.metadata?.comment }}</CTableDataCell>
                            <CTableDataCell>
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
                        </template>
                    </CTableRow>
                </CTableBody>
            </CTable>
        </CCardBody>
        <CCardFooter v-if="paginationMetadata.last_page > 1" class="d-flex justify-content-center">
            <Paginator :current-page="currentPage" :total-pages="paginationMetadata.last_page"
                @page-changed="onPageChanged" />
        </CCardFooter>
    </CCard>

    <CModal :visible="deletingId !== null" @close="deletingId = null">
        <CModalHeader>
            <CModalTitle>Delete Protected Address</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to delete this protected address? This action cannot be undone.</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader>
            <CModalTitle>Deactivate Protected Address</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to deactivate this protected address? Aliases forwarding to it will stop
            delivering email.</CModalBody>
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
        <CModalBody>Are you sure you want to activate this protected address?</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingActivateId = null">Cancel</CButton>
            <CButton color="success" :disabled="saving" @click="setActive(confirmingActivateId, true)">Yes, activate
            </CButton>
        </CModalFooter>
    </CModal>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-clicked'])
const praddrs = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const searchQuery = ref('')
const showSearch = ref(false)
const searchInputRef = ref(null)
const editingId = ref(null)
const editComment = ref('')
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)

let searchDebounce = null

const load = async () => {
    const params = new URLSearchParams({ page: currentPage.value })
    if (searchQuery.value) params.set('q', searchQuery.value)
    const res = await apiFetch('/api/v1/praddrs?' + params)
    const data = await res.json()
    praddrs.value = data.protected_addresses
    paginationMetadata.value = data.pagination_metadata
}

watch(searchQuery, () => {
    clearTimeout(searchDebounce)
    searchDebounce = setTimeout(() => {
        currentPage.value = 1
        load()
    }, 300)
})

const openSearch = async () => {
    showSearch.value = true
    await nextTick()
    searchInputRef.value?.$el?.focus()
}

const closeSearch = () => {
    showSearch.value = false
    searchQuery.value = ''
}

const startEdit = (addr) => {
    editingId.value = addr.id
    editComment.value = addr.metadata?.comment ?? ''
}

const cancelEdit = () => {
    editingId.value = null
}

const saveEdit = async (id) => {
    saving.value = true
    await apiFetch(`/api/v1/praddrs/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({
            metadata: { comment: editComment.value },
        }),
    })
    saving.value = false
    editingId.value = null
    await load()
}

const setActive = async (id, active) => {
    saving.value = true
    await apiFetch(`/api/v1/praddrs/${id}`, {
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
    await apiFetch(`/api/v1/praddrs/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
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
