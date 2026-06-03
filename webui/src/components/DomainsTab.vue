<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Domains</span>
            <CButton color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Name</CTableHeaderCell>
                        <CTableHeaderCell>Type</CTableHeaderCell>
                        <CTableHeaderCell>Status</CTableHeaderCell>
                        <CTableHeaderCell v-if="props.userInfo.type === 'admin'">Owner</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="domain in domains" :key="domain.id">
                        <CTableDataCell>{{ domain.name }}</CTableDataCell>
                        <CTableDataCell>
                            <CBadge :color="domain.type === 'global' ? 'info' : 'secondary'">
                                {{ domain.type }}
                            </CBadge>
                        </CTableDataCell>
                        <CTableDataCell>
                            <CBadge :color="domain.active ? 'success' : 'danger'">
                                {{ domain.active ? 'Active' : 'Inactive' }}
                            </CBadge>
                        </CTableDataCell>
                        <CTableDataCell v-if="props.userInfo.type === 'admin'">
                            {{ domain.owner?.login ?? '—' }}
                        </CTableDataCell>
                        <CTableDataCell class="text-end text-nowrap">
                            <CButton v-if="domain.active" v-c-tooltip="'Deactivate'" color="warning" size="sm"
                                variant="outline" class="me-1" @click="confirmingDeactivateId = domain.id">
                                <CIcon icon="cilBan" />
                            </CButton>
                            <CButton v-else v-c-tooltip="'Activate'" color="success" size="sm" variant="outline"
                                class="me-1" @click="confirmingActivateId = domain.id">
                                <CIcon icon="cilCheckCircle" />
                            </CButton>
                            <CButton v-c-tooltip="'Delete'" color="danger" size="sm" variant="outline"
                                @click="deletingId = domain.id">
                                <CIcon icon="cilTrash" />
                            </CButton>
                        </CTableDataCell>
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
            <CModalTitle>Delete Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to delete this domain? This action cannot be undone.</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader>
            <CModalTitle>Deactivate Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to deactivate this domain?</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingDeactivateId = null">Cancel</CButton>
            <CButton color="warning" :disabled="saving" @click="setActive(confirmingDeactivateId, false)">Yes,
                deactivate
            </CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingActivateId !== null" @close="confirmingActivateId = null">
        <CModalHeader>
            <CModalTitle>Activate Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to activate this domain?</CModalBody>
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
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['add-clicked'])

const domains = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)
const apiError = ref(null)

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

const load = async () => {
    const params = new URLSearchParams({ page: currentPage.value, include_global: 'true' })
    const res = await apiFetch('/api/v1/domains?' + params)
    const data = await res.json()
    domains.value = data.domains
    paginationMetadata.value = data.pagination_metadata ?? {}
}

const setActive = async (id, active) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/domains/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active }),
    })
    saving.value = false
    confirmingDeactivateId.value = null
    confirmingActivateId.value = null
    if (!res.ok) { await handleApiError(res); return }
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/domains/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
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
