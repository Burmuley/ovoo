<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button @click="addPrAddr">Add new Protected address</button>
        </div>
        <div v-for="(addr, index) in praddrs" :key="addr.id" class="ovoo-item" :class="{ dark: index % 2 !== 0 }">
            <div class="ovoo-item-content">
                <p>{{ addr.email }}</p>
                <p v-if="addr.metadata.comment"><small>Comment: {{ addr.metadata.comment }}</small></p>
            </div>
            <div class="ovoo-item buttons" :class="{ dark: index % 2 !== 0 }">
                <!-- <button @click=" edit(addr)" style="margin-right: 5px;">Edit</button> -->
                <button @click="remove(addr.id)">Delete</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-praddr-clicked'])

const praddrs = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/praddrs')
    praddrs.value = await res.json()
}

const edit = (addr) => {
    console.log('edit not implemented', addr)
}

const remove = async (id) => {
    await apiFetch(`/api/v1/praddrs/${id}`, { method: 'DELETE' })
    await load()
}

const addPrAddr = () => {
    emit('add-praddr-clicked')
}

onMounted(load)
</script>
