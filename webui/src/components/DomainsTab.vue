<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <div class="d-flex align-items-center">
                <span class="fw-semibold">Domains</span>
                <InfoPopover
                    description="Domains define the address space for creating aliases (e.g. @example.com). Global domains are available to all users; personal domains are owned by individual users." />
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
                <CButton color="primary" size="sm" @click="emit('add-clicked')">
                    <CIcon icon="cilPlus" /> Add
                </CButton>
            </div>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Name</CTableHeaderCell>
                        <CTableHeaderCell class="text-center" style="width: 1%; white-space: nowrap;">Type
                        </CTableHeaderCell>
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
                    <EmptyState v-else-if="domains.length === 0" icon="cilGlobeAlt"
                        message="No domains yet. Add one to use custom addresses for aliases." action-label="Add Domain"
                        :colspan="4" @action-clicked="emit('add-clicked')" />
                    <template v-else>
                        <CTableRow v-for="domain in domains" :key="domain.id">
                            <CTableDataCell>
                                <div>{{ domain.name }}</div>
                                <div v-if="props.userInfo.type === 'admin'" class="text-body-secondary"
                                    style="font-size: 0.75rem;">
                                    Owner: {{ domain.owner?.login ?? '—' }}
                                </div>
                            </CTableDataCell>
                            <CTableDataCell class="text-center text-nowrap">
                                <CBadge :color="domain.type === 'global' ? 'info' : 'secondary'">
                                    {{ domain.type }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-center text-nowrap">
                                <CBadge :color="domain.active ? 'success' : 'danger'">
                                    {{ domain.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                                <CBadge v-if="domain.type === 'personal'"
                                    :color="domain.verified ? 'success' : 'warning'" class="ms-1">
                                    {{ domain.verified ? 'Verified' : 'Unverified' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton v-if="domain.type === 'personal' && !domain.verified" v-c-tooltip="'Verify'"
                                    color="primary" size="sm" variant="outline" class="me-1"
                                    @click="openVerify(domain)">
                                    <CIcon icon="cilShieldAlt" />
                                </CButton>
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
                    </template>
                </CTableBody>
            </CTable>
        </CCardBody>
        <CCardFooter v-if="paginationMetadata.last_page > 1" class="d-flex justify-content-center">
            <Paginator :current-page="currentPage" :total-pages="paginationMetadata.last_page"
                :total-items="paginationMetadata.total_records" @page-changed="onPageChanged" />
        </CCardFooter>
    </CCard>

    <CModal :visible="deletingId !== null" @close="deletingId = null">
        <CModalHeader>
            <CModalTitle>Delete Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Delete domain <strong>{{ deletingDomain?.name }}</strong>? This action cannot be undone.
        </CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader>
            <CModalTitle>Deactivate Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Deactivate <strong>{{ confirmingDeactivateDomain?.name }}</strong>?
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
            <CModalTitle>Activate Domain</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Activate <strong>{{ confirmingActivateDomain?.name }}</strong>?
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

    <VerifyDomainModal :domain="verifyingDomain" @close="verifyingDomain = null" @verify-complete="load" />
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'
import Paginator from './Paginator.vue'
import EmptyState from './EmptyState.vue'
import InfoPopover from './InfoPopover.vue'
import VerifyDomainModal from './VerifyDomainModal.vue'

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['add-clicked'])
const { showToast } = useToast()

const domains = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const searchQuery = ref('')
const loading = ref(true)
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)
const apiError = ref(null)
const verifyingDomain = ref(null)

const deletingDomain = computed(() => domains.value.find(d => d.id === deletingId.value))
const confirmingDeactivateDomain = computed(() => domains.value.find(d => d.id === confirmingDeactivateId.value))
const confirmingActivateDomain = computed(() => domains.value.find(d => d.id === confirmingActivateId.value))

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

let searchDebounce = null

const load = async () => {
    loading.value = true
    const params = new URLSearchParams({ page: currentPage.value, include_global: 'true' })
    if (searchQuery.value) params.set('domain_name', searchQuery.value)
    const res = await apiFetch('/api/v1/domains?' + params)
    const data = await res.json()
    domains.value = data.domains
    paginationMetadata.value = data.pagination_metadata ?? {}
    loading.value = false
}

watch(searchQuery, () => {
    clearTimeout(searchDebounce)
    searchDebounce = setTimeout(() => {
        currentPage.value = 1
        load()
    }, 300)
})

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
    showToast(active ? 'Domain activated.' : 'Domain deactivated.')
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/domains/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('Domain deleted.')
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

const openVerify = (domain) => { verifyingDomain.value = domain }

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>
