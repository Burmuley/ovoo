<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button title="Add new Alias" @click="addAlias">+</button>
            <Paginator v-if="paginationMetadata.last_page > 1" current_page="1"
                :total_pages="paginationMetadata.last_page" @page-changed=onPageChanged />
        </div>
        <div v-for="(alias, index) in aliases" :key="alias.id" class="ovoo-item" :class="{ dark: index % 2 !== 0 }">
            <div class="ovoo-item-content">
                <p>{{ alias.email }}</p>
                <p><small>Forwards to: {{ alias.forward_email }}</small></p>
                <p v-if="alias.metadata.service_name"><small>Linked service: {{ alias.metadata.service_name }}</small>
                </p>
                <p v-if="alias.metadata.comment"><small>Comment: {{ alias.metadata.comment }}</small>
                </p>
            </div>
            <div class="ovoo-item buttons" :class="{ dark: index % 2 !== 0 }">
                <!-- <button @click="edit(alias)" title="Edit alias" style="margin-right: 5px;">Edit</button> -->
                <button @click="deleteAlias(alias.id)" title="Delete alias">&#9932;</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-alias-clicked'])
const data = ref({})
const aliases = ref([])
const paginationMetadata = ref({})
const currentPage = ref(1)

const load = async () => {
    const res = await apiFetch('/api/v1/aliases?page=' + currentPage.value)
    data.value = await res.json()
    aliases.value = data.value.aliases
    paginationMetadata.value = data.value.pagination_metadata
}

const edit = (alias) => {
    console.log('edit not implemented', alias)
}

const deleteAlias = async (id) => {
    await apiFetch(`/api/v1/aliases/${id}`, { method: 'DELETE' })
    await load()
}

const addAlias = () => {
    emit('add-alias-clicked')
}

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}

onMounted(load)
</script>
