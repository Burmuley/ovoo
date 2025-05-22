<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button @click="addApiKey">Add new API key</button>

        </div>
        <div v-for="(api_key, index) in api_keys" :key="api_key.id" class="ovoo-item"
            :class="{ dark: index % 2 !== 0 }">
            <div class="ovoo-item-content">
                <p><b>{{ api_key.name }}</b></p>
                <p>Description: '{{ api_key.description }}'</p>
                <p>Expiration: {{ moment(api_key.expiration).format('LLL') }}</p>
            </div>
            <div class="ovoo-item buttons" :class="{ dark: index % 2 !== 0 }">
                <button @click="deleteApiKey(api_key.id)" title="Delete API key">Delete</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import moment from 'moment'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-apikey-clicked'])
const api_keys = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/users/apitokens')
    api_keys.value = await res.json()
}

const deleteApiKey = async (id) => {
    await apiFetch(`/api/v1/users/apitokens/${id}`, { method: 'DELETE' })
    await load()
}

const addApiKey = () => {
    emit('add-apikey-clicked')
}

onMounted(load)
</script>
