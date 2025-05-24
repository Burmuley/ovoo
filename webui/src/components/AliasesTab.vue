<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button @click="addAlias">Add new Alias</button>

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
                <button @click="deleteAlias(alias.id)" title="Delete alias">Delete</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-alias-clicked'])
const aliases = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/aliases')
    aliases.value = await res.json()
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

onMounted(load)
</script>
