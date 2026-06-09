<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <div class="d-flex align-items-center">
                <span class="fw-semibold">Users</span>
                <InfoPopover
                    description="User accounts control who can log in to Ovoo. Admins can create and manage all accounts. Regular users manage only their own aliases and protected addresses. Milter users are service accounts used by the mail filter integration." />
            </div>
            <CButton v-if="props.userInfo.type === 'admin'" color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
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
                    <EmptyState v-else-if="users.length === 0" icon="cilPeople" message="No users found."
                        :action-label="props.userInfo.type === 'admin' ? 'Add User' : ''" :colspan="4"
                        @action-clicked="emit('add-clicked')" />
                    <template v-else>
                        <CTableRow v-for="user in users" :key="user.id" style="cursor: pointer"
                            @click="selectedUser = user">
                            <CTableDataCell>
                                <div>{{ user.first_name }} {{ user.last_name }}</div>
                                <div class="text-body-secondary" style="font-size: 0.75rem;">{{ user.login }}</div>
                            </CTableDataCell>
                            <CTableDataCell class="text-center text-nowrap">
                                <CBadge :color="typeBadgeColor(user.type)">{{ user.type }}</CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-center text-nowrap">
                                <CBadge :color="user.active ? 'success' : 'danger'">
                                    {{ user.active ? 'Active' : 'Inactive' }}
                                </CBadge>
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <template v-if="props.userInfo.type === 'admin'">
                                    <CButton v-if="user.active" color="warning" size="sm" variant="outline" class="me-1"
                                        @click.stop="confirmingDeactivateId = user.id">
                                        <CIcon icon="cilBan" />
                                    </CButton>
                                    <CButton v-else color="success" size="sm" variant="outline" class="me-1"
                                        @click.stop="confirmingActivateId = user.id">
                                        <CIcon icon="cilCheckCircle" />
                                    </CButton>
                                    <CButton color="danger" size="sm" variant="outline"
                                        @click.stop="deletingId = user.id">
                                        <CIcon icon="cilTrash" />
                                    </CButton>
                                </template>
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

    <CModal :visible="selectedUser !== null" @close="handleModalClose">
        <CModalHeader>
            <CModalTitle>{{ selectedUser?.first_name }} {{ selectedUser?.last_name }}</CModalTitle>
        </CModalHeader>
        <CModalBody>
            <table class="table table-sm table-borderless mb-0">
                <tbody>
                    <tr>
                        <th>ID</th>
                        <td class="text-break">{{ selectedUser?.id }}</td>
                    </tr>
                    <tr>
                        <th>Login</th>
                        <td>{{ selectedUser?.login }}</td>
                    </tr>
                    <tr>
                        <th>First Name</th>
                        <td>
                            <div v-if="editingField === 'first_name'" class="d-flex gap-1 align-items-center">
                                <CFormInput size="sm" v-model="editValue" autofocus class="flex-grow-1"
                                    @keydown.enter.prevent="saveEdit" />
                                <CButton size="sm" color="success" variant="outline" :disabled="saving"
                                    @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving"
                                    @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else class="d-flex align-items-center justify-content-between gap-2">
                                <span>{{ selectedUser?.first_name }}</span>
                                <CButton size="sm" color="primary" variant="outline"
                                    @click="startEdit('first_name', selectedUser.first_name)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <th>Last Name</th>
                        <td>
                            <div v-if="editingField === 'last_name'" class="d-flex gap-1 align-items-center">
                                <CFormInput size="sm" v-model="editValue" autofocus class="flex-grow-1"
                                    @keydown.enter.prevent="saveEdit" />
                                <CButton size="sm" color="success" variant="outline" :disabled="saving"
                                    @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving"
                                    @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else class="d-flex align-items-center justify-content-between gap-2">
                                <span>{{ selectedUser?.last_name }}</span>
                                <CButton size="sm" color="primary" variant="outline"
                                    @click="startEdit('last_name', selectedUser.last_name)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <th>Type</th>
                        <td>
                            <div v-if="editingField === 'type'" class="d-flex gap-1 align-items-center">
                                <CFormSelect size="sm" v-model="editValue" class="flex-grow-1"
                                    @keydown.enter.prevent="saveEdit">
                                    <option value="regular">regular</option>
                                    <option value="admin">admin</option>
                                    <option value="milter">milter</option>
                                </CFormSelect>
                                <CButton size="sm" color="success" variant="outline" :disabled="saving"
                                    @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving"
                                    @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else-if="props.userInfo.type === 'admin'"
                                class="d-flex align-items-center justify-content-between gap-2">
                                <CBadge :color="typeBadgeColor(selectedUser?.type)">{{ selectedUser?.type }}</CBadge>
                                <CButton size="sm" color="primary" variant="outline"
                                    @click="startEdit('type', selectedUser.type)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                            <CBadge v-else :color="typeBadgeColor(selectedUser?.type)">{{ selectedUser?.type }}</CBadge>
                        </td>
                    </tr>
                    <tr>
                        <th>Failed Attempts</th>
                        <td>{{ selectedUser?.failed_attempts ?? 0 }}</td>
                    </tr>
                    <tr>
                        <th>Locked Until</th>
                        <td>{{ selectedUser?.lockout_until ? moment(selectedUser.lockout_until).format('LLL') : '—' }}
                        </td>
                    </tr>
                </tbody>
            </table>
        </CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="selectedUser = null">Close</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="deletingId !== null" @close="deletingId = null">
        <CModalHeader>
            <CModalTitle>Delete User</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Delete user <strong>{{ deletingUser?.login }}</strong>? This action cannot be undone.
        </CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader>
            <CModalTitle>Deactivate User</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Deactivate <strong>{{ confirmingDeactivateUser?.login }}</strong>? They will no longer be able to log in.
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
            <CModalTitle>Activate User</CModalTitle>
        </CModalHeader>
        <CModalBody>
            Activate <strong>{{ confirmingActivateUser?.login }}</strong>?
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
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import moment from 'moment'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'
import Paginator from './Paginator.vue'
import EmptyState from './EmptyState.vue'
import InfoPopover from './InfoPopover.vue'

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['add-clicked'])
const { showToast } = useToast()

const users = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const loading = ref(true)
const selectedUser = ref(null)
const editingField = ref(null)
const editValue = ref('')
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)
const apiError = ref(null)

const deletingUser = computed(() => users.value.find(u => u.id === deletingId.value))
const confirmingDeactivateUser = computed(() => users.value.find(u => u.id === confirmingDeactivateId.value))
const confirmingActivateUser = computed(() => users.value.find(u => u.id === confirmingActivateId.value))

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

watch(selectedUser, () => {
    editingField.value = null
    editValue.value = ''
})

const typeBadgeColor = (type) => {
    if (type === 'admin') return 'danger'
    if (type === 'milter') return 'warning'
    return 'secondary'
}

function startEdit(field, value) {
    editingField.value = field
    editValue.value = value
}

async function saveEdit() {
    if (!editingField.value || saving.value) return
    saving.value = true
    const res = await apiFetch(`/api/v1/users/${selectedUser.value.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ [editingField.value]: editValue.value }),
    })
    saving.value = false
    if (!res.ok) { await handleApiError(res); return }
    selectedUser.value = { ...selectedUser.value, [editingField.value]: editValue.value }
    editingField.value = null
    showToast('User updated.')
    await load()
}

function cancelEdit() {
    editingField.value = null
}

function handleModalClose() {
    if (editingField.value !== null) {
        cancelEdit()
    } else {
        selectedUser.value = null
    }
}

const load = async () => {
    loading.value = true
    const res = await apiFetch('/api/v1/users?page=' + currentPage.value)
    const data = await res.json()
    users.value = data.users
    paginationMetadata.value = data.pagination_metadata
    loading.value = false
}

const setActive = async (id, active) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/users/${id}`, {
        method: 'PATCH',
        body: JSON.stringify({ active }),
    })
    saving.value = false
    confirmingDeactivateId.value = null
    confirmingActivateId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast(active ? 'User activated.' : 'User deactivated.')
    await load()
}

const performDelete = async (id) => {
    saving.value = true
    const res = await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' })
    saving.value = false
    deletingId.value = null
    if (!res.ok) { await handleApiError(res); return }
    showToast('User deleted.')
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
