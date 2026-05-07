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
                    <CTableRow v-for="user in users" :key="user.id">
                        <CTableDataCell>{{ user.first_name }} {{ user.last_name }}</CTableDataCell>
                        <CTableDataCell>{{ user.login }}</CTableDataCell>
                        <CTableDataCell>
                            <CBadge :color="typeBadgeColor(user.type)">{{ user.type }}</CBadge>
                        </CTableDataCell>
                        <CTableDataCell class="text-end">
                            <CButton color="danger" size="sm" variant="outline" @click="deleteUser(user.id)">
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
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-clicked'])
const users = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)

const typeBadgeColor = (type) => {
    if (type === 'admin') return 'danger'
    if (type === 'milter') return 'warning'
    return 'secondary'
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
