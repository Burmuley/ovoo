<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Users</span>
            <CButton color="primary" size="sm" @click="emit('add-clicked')">
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
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow
                        v-for="user in users"
                        :key="user.id"
                        style="cursor: pointer"
                        @click="selectedUser = user"
                    >
                        <CTableDataCell>{{ user.first_name }} {{ user.last_name }}</CTableDataCell>
                        <CTableDataCell>{{ user.login }}</CTableDataCell>
                        <CTableDataCell>
                            <CBadge :color="typeBadgeColor(user.type)">{{ user.type }}</CBadge>
                        </CTableDataCell>
                        <CTableDataCell class="text-end">
                            <CButton color="danger" size="sm" variant="outline" @click.stop="deleteUser(user.id)">
                                <CIcon icon="cilTrash" />
                            </CButton>
                        </CTableDataCell>
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

    <CModal :visible="selectedUser !== null" @close="handleModalClose">
        <CModalHeader>
            <CModalTitle>User Details</CModalTitle>
        </CModalHeader>
        <CModalBody>
            <table class="table table-sm table-borderless mb-0">
                <tbody>
                    <tr><th>ID</th><td class="text-break">{{ selectedUser?.id }}</td></tr>
                    <tr><th>Login</th><td>{{ selectedUser?.login }}</td></tr>
                    <tr>
                        <th>First Name</th>
                        <td>
                            <div v-if="editingField === 'first_name'" class="d-flex gap-1 align-items-center">
                                <CFormInput
                                    size="sm"
                                    v-model="editValue"
                                    autofocus
                                    class="flex-grow-1"
                                    @keydown.enter.prevent="saveEdit"
                                />
                                <CButton size="sm" color="success" variant="outline" :disabled="saving" @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving" @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else class="d-flex align-items-center justify-content-between gap-2">
                                <span>{{ selectedUser?.first_name }}</span>
                                <CButton size="sm" color="primary" variant="outline" @click="startEdit('first_name', selectedUser.first_name)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <th>Last Name</th>
                        <td>
                            <div v-if="editingField === 'last_name'" class="d-flex gap-1 align-items-center">
                                <CFormInput
                                    size="sm"
                                    v-model="editValue"
                                    autofocus
                                    class="flex-grow-1"
                                    @keydown.enter.prevent="saveEdit"
                                />
                                <CButton size="sm" color="success" variant="outline" :disabled="saving" @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving" @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else class="d-flex align-items-center justify-content-between gap-2">
                                <span>{{ selectedUser?.last_name }}</span>
                                <CButton size="sm" color="primary" variant="outline" @click="startEdit('last_name', selectedUser.last_name)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                        </td>
                    </tr>
                    <tr>
                        <th>Type</th>
                        <td>
                            <div v-if="editingField === 'type'" class="d-flex gap-1 align-items-center">
                                <CFormSelect size="sm" v-model="editValue" class="flex-grow-1" @keydown.enter.prevent="saveEdit">
                                    <option value="regular">regular</option>
                                    <option value="admin">admin</option>
                                    <option value="milter">milter</option>
                                </CFormSelect>
                                <CButton size="sm" color="success" variant="outline" :disabled="saving" @click="saveEdit">
                                    <CIcon icon="cilCheck" />
                                </CButton>
                                <CButton size="sm" color="secondary" variant="outline" :disabled="saving" @click="cancelEdit">
                                    <CIcon icon="cilX" />
                                </CButton>
                            </div>
                            <div v-else-if="props.userInfo.type === 'admin'" class="d-flex align-items-center justify-content-between gap-2">
                                <CBadge :color="typeBadgeColor(selectedUser?.type)">{{ selectedUser?.type }}</CBadge>
                                <CButton size="sm" color="primary" variant="outline" @click="startEdit('type', selectedUser.type)">
                                    <CIcon icon="cilPencil" />
                                </CButton>
                            </div>
                            <CBadge v-else :color="typeBadgeColor(selectedUser?.type)">{{ selectedUser?.type }}</CBadge>
                        </td>
                    </tr>
                    <tr><th>Failed Attempts</th><td>{{ selectedUser?.failed_attempts ?? 0 }}</td></tr>
                    <tr><th>Locked Until</th><td>{{ selectedUser?.lockout_until || '—' }}</td></tr>
                </tbody>
            </table>
        </CModalBody>
        <CModalFooter>
            <CButton color="secondary" variant="outline" @click="selectedUser = null">Close</CButton>
        </CModalFooter>
    </CModal>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
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

const deleteUser = async (id) => {
    await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' })
    await load()
}

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>

