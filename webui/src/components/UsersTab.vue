<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Users</span>
            <CButton v-if="props.userInfo.type === 'admin'" color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Name</CTableHeaderCell>
                        <CTableHeaderCell>Login</CTableHeaderCell>
                        <CTableHeaderCell>Type</CTableHeaderCell>
                        <CTableHeaderCell>Status</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="user in users" :key="user.id" style="cursor: pointer"
                        @click="selectedUser = user">
                        <CTableDataCell>{{ user.first_name }} {{ user.last_name }}</CTableDataCell>
                        <CTableDataCell>{{ user.login }}</CTableDataCell>
                        <CTableDataCell>
                            <CBadge :color="typeBadgeColor(user.type)">{{ user.type }}</CBadge>
                        </CTableDataCell>
                        <CTableDataCell>
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
                                <CButton color="danger" size="sm" variant="outline" @click.stop="deletingId = user.id">
                                    <CIcon icon="cilTrash" />
                                </CButton>
                            </template>
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

    <CModal :visible="selectedUser !== null" @close="handleModalClose">
        <CModalHeader>
            <CModalTitle>User Details</CModalTitle>
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
                        <td>{{ selectedUser?.lockout_until || '—' }}</td>
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
        <CModalBody>Are you sure you want to delete this user? This action cannot be undone.</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="deletingId = null">Cancel</CButton>
            <CButton color="danger" :disabled="saving" @click="performDelete(deletingId)">Yes, delete</CButton>
        </CModalFooter>
    </CModal>

    <CModal :visible="confirmingDeactivateId !== null" @close="confirmingDeactivateId = null">
        <CModalHeader>
            <CModalTitle>Deactivate User</CModalTitle>
        </CModalHeader>
        <CModalBody>Are you sure you want to deactivate this user? They will no longer be able to log in.</CModalBody>
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
        <CModalBody>Are you sure you want to activate this user?</CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="confirmingActivateId = null">Cancel</CButton>
            <CButton color="success" :disabled="saving" @click="setActive(confirmingActivateId, true)">Yes, activate
            </CButton>
        </CModalFooter>
    </CModal>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const props = defineProps({ userInfo: { type: Object, default: () => ({}) } })
const emit = defineEmits(['add-clicked'])

const users = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const selectedUser = ref(null)
const editingField = ref(null)
const editValue = ref('')
const saving = ref(false)
const deletingId = ref(null)
const confirmingDeactivateId = ref(null)
const confirmingActivateId = ref(null)

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
    await apiFetch(`/api/v1/users/${selectedUser.value.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ [editingField.value]: editValue.value }),
    })
    selectedUser.value = { ...selectedUser.value, [editingField.value]: editValue.value }
    editingField.value = null
    saving.value = false
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
    const res = await apiFetch('/api/v1/users?page=' + currentPage.value)
    const data = await res.json()
    users.value = data.users
    paginationMetadata.value = data.pagination_metadata
}

const setActive = async (id, active) => {
    saving.value = true
    await apiFetch(`/api/v1/users/${id}`, {
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
    await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' })
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
