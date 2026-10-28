<template>
<CModal :visible="alias !== null" size="lg" @close="emit('close')">
    <CModalHeader>
        <CModalTitle class="d-flex align-items-center">
            <span>Chains — {{ alias?.email }}</span>
            <InfoPopover
                description="A chain links one outside correspondent to this alias. Normally a chain is created when they email your alias first. Start one here to write first yourself: Ovoo generates a private address, and anything you send to it from your protected address reaches the recipient as if it came from the alias." />
        </CModalTitle>
    </CModalHeader>
    <CModalBody>
        <CAlert v-if="created" color="success" class="mb-3">
            <div class="mb-2">Chain created. Send your email <strong>from {{ created.to_email }}</strong> to the
                address below — {{ created.orig_from_address.email }} will see it coming from {{ alias?.email }}.</div>
            <div class="d-flex align-items-start justify-content-between gap-2">
                <code style="min-width: 0; word-break: break-all;">{{ created.from_email }}</code>
                <CButton size="sm" color="success" variant="outline" class="flex-shrink-0"
                    @click="copy(created.from_email, 'new')">
                    {{ copiedKey === 'new' ? 'Copied!' : 'Copy' }}
                </CButton>
            </div>
        </CAlert>

        <CAlert v-if="apiError" color="warning" class="mb-3" style="word-break: break-word;">
            {{ apiError }}
        </CAlert>

        <CTable hover responsive class="mb-3">
            <CTableHead>
                <CTableRow>
                    <CTableHeaderCell>Recipient</CTableHeaderCell>
                    <CTableHeaderCell>Send to</CTableHeaderCell>
                    <CTableHeaderCell style="width: 1%; white-space: nowrap;">Started</CTableHeaderCell>
                    <CTableHeaderCell style="width: 1%;"></CTableHeaderCell>
                </CTableRow>
            </CTableHead>
            <CTableBody>
                <template v-if="loading">
                    <CTableRow v-for="n in 2" :key="n">
                        <CTableDataCell v-for="c in 4" :key="c">
                            <div class="placeholder-glow"><span class="placeholder col-8"></span></div>
                        </CTableDataCell>
                    </CTableRow>
                </template>
                <CTableRow v-else-if="chains.length === 0">
                    <CTableDataCell colspan="4" class="text-center text-body-secondary py-4">
                        No chains yet. Start one below.
                    </CTableDataCell>
                </CTableRow>
                <template v-else>
                    <CTableRow v-for="chain in chains" :key="chain.hash">
                        <CTableDataCell>
                            <div v-c-tooltip="chain.orig_from_address.email" class="text-truncate"
                                style="max-width: 200px;">{{ chain.orig_from_address.email }}</div>
                        </CTableDataCell>
                        <CTableDataCell>
                            <div v-c-tooltip="chain.from_email" class="text-truncate" style="max-width: 260px;">
                                <code>{{ chain.from_email }}</code>
                            </div>
                        </CTableDataCell>
                        <CTableDataCell class="text-nowrap text-body-secondary" style="font-size: 0.8rem;">
                            {{ formatDate(chain.created_at) }}
                        </CTableDataCell>
                        <CTableDataCell class="text-end">
                            <CButton v-c-tooltip="'Copy address'" size="sm" color="secondary" variant="outline"
                                @click="copy(chain.from_email, chain.hash)">
                                {{ copiedKey === chain.hash ? 'Copied!' : 'Copy' }}
                            </CButton>
                        </CTableDataCell>
                    </CTableRow>
                </template>
            </CTableBody>
        </CTable>

        <CFormLabel>Start a new chain</CFormLabel>
        <div class="input-group">
            <CFormInput v-model="recipient" type="email" placeholder="recipient@example.com" :disabled="creating"
                @keyup.enter="createChain" />
            <CButton color="primary" :disabled="creating || !recipient.trim()" @click="createChain">
                <CSpinner v-if="creating" size="sm" class="me-1" />Start
            </CButton>
        </div>
        <div class="form-text">The recipient's real address. They never learn your protected address.</div>
    </CModalBody>
    <CModalFooter>
        <CButton color="secondary" variant="outline" @click="emit('close')">Close</CButton>
    </CModalFooter>
</CModal>
</template>

<script setup>
import { ref, watch } from 'vue'
import { apiFetch } from '../utils/api'
import { useToast } from '../composables/useToast'
import InfoPopover from './InfoPopover.vue'

const props = defineProps({ alias: { type: Object, default: null } })
const emit = defineEmits(['close'])
const { showToast } = useToast()

const chains = ref([])
const loading = ref(false)
const creating = ref(false)
const recipient = ref('')
const created = ref(null)
const apiError = ref(null)
const copiedKey = ref(null)

const formatDate = (value) => (value ? new Date(value).toLocaleDateString() : '')

const handleApiError = async (res) => {
    const data = await res.json()
    apiError.value = data.errors?.[0]?.detail ?? 'An unexpected error occurred'
}

const load = async () => {
    loading.value = true
    const res = await apiFetch(`/api/v1/chains?alias=${props.alias.id}`)
    loading.value = false
    if (!res.ok) { await handleApiError(res); return }
    const data = await res.json()
    chains.value = data.chains
}

watch(() => props.alias, (val) => {
    chains.value = []
    recipient.value = ''
    created.value = null
    apiError.value = null
    copiedKey.value = null
    if (val !== null) load()
})

const copy = async (text, key) => {
    await navigator.clipboard.writeText(text)
    copiedKey.value = key
    setTimeout(() => { if (copiedKey.value === key) copiedKey.value = null }, 2000)
}

const createChain = async () => {
    const to = recipient.value.trim()
    if (!to) return

    creating.value = true
    apiError.value = null
    const res = await apiFetch('/api/v1/chains', {
        method: 'POST',
        body: JSON.stringify({ from_email: to, to_email: props.alias.email }),
    })
    creating.value = false
    if (!res.ok) { await handleApiError(res); return }

    created.value = await res.json()
    recipient.value = ''
    showToast('Chain created.')
    await load()
}
</script>
