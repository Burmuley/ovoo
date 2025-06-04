<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button @click="addUser">Add new user</button>
            <Paginator v-if="paginationMetadata.last_page > 1" current_page="1"
                :total_pages="paginationMetadata.last_page" @page-changed=onPageChanged />
        </div>
        <div v-for="(user, index) in users" :key="user.id" class="ovoo-item" :class="{ dark: index % 2 !== 0 }">
            <div class="ovoo-item-content">
                <p>{{ user.first_name }} {{ user.last_name }} ({{ user.login }})</p>
                <p v-if="user.type"><small>Type: {{ user.type }}</small></p>
            </div>
            <div class="ovoo-item buttons" :class="{ dark: index % 2 !== 0 }">
                <!-- <button @click="edit(user)" title="Edit user" style="margin-right: 5px;">Edit</button> -->
                <button @click="deleteUser(user.id)" title="Delete user">Delete</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import Paginator from './Paginator.vue'

const emit = defineEmits(['add-user-clicked'])
const users = ref([])
const data = ref({})
const paginationMetadata = ref({})
const currentPage = ref(1)

const load = async () => {
    const res = await apiFetch('/api/v1/users?&page=' + currentPage.value)
    data.value = await res.json()
    users.value = data.value.users
    paginationMetadata.value = data.value.pagination_metadata
}

const edit = (alias) => {
    console.log('edit not implemented', alias)
}

const deleteUser = async (id) => {
    await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' })
    await load()
}

const addUser = () => {
    emit('add-user-clicked')
}

const onPageChanged = async (page) => {
    currentPage.value = page
    await load()
}


onMounted(load)
</script>
