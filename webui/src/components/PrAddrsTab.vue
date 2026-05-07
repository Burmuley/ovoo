<template>
    <CCard>
        <CCardHeader class="d-flex align-items-center justify-content-between">
            <span class="fw-semibold">Protected Addresses</span>
            <CButton color="primary" size="sm" @click="emit('add-clicked')">
                <CIcon icon="cilPlus" /> Add
            </CButton>
        </CCardHeader>
        <CCardBody class="p-0">
            <CTable hover responsive class="mb-0">
                <CTableHead>
                    <CTableRow>
                        <CTableHeaderCell>Email</CTableHeaderCell>
                        <CTableHeaderCell>Comment</CTableHeaderCell>
                        <CTableHeaderCell></CTableHeaderCell>
                    </CTableRow>
                </CTableHead>
                <CTableBody>
                    <CTableRow v-for="addr in praddrs" :key="addr.id">
                        <CTableDataCell>{{ addr.email }}</CTableDataCell>

                        <template v-if="editingId === addr.id">
                            <CTableDataCell>
                                <CFormInput
                                    v-model="editComment"
                                    size="sm"
                                    placeholder="Comment"
                                    @keyup.enter="saveEdit(addr.id)"
                                />
                            </CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton
                                    color="success"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    :disabled="saving"
                                    @click="saveEdit(addr.id)"
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
                            <CTableDataCell>{{ addr.metadata?.comment }}</CTableDataCell>
                            <CTableDataCell class="text-end text-nowrap">
                                <CButton
                                    color="primary"
                                    size="sm"
                                    variant="outline"
                                    class="me-1"
                                    @click="startEdit(addr)"
                                >
                                    <CIcon icon="cilPencil" />
                                </CButton>
                                <CButton
                                    color="danger"
                                    size="sm"
                                    variant="outline"
                                    @click="remove(addr.id)"
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
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-clicked'])
const praddrs = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)
const editingId = ref(null)
const editComment = ref('')
const saving = ref(false)

const load = async () => {
    const res = await apiFetch('/api/v1/praddrs?page=' + currentPage.value)
    const data = await res.json()
    praddrs.value = data.protected_addresses
    paginationMetadata.value = data.pagination_metadata
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

const remove = async (id) => {
    await apiFetch(`/api/v1/praddrs/${id}`, { method: 'DELETE' })
    await load()
}

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>
