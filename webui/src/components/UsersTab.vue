<template>
    <div class="ovoo-items-list">
        <div class="ovoo-item header">
            <button @click="addUser">Add new User</button>

        </div>
        <div v-for="(user, index) in users" :key="user.id" class="ovoo-item" :class="{ dark: index % 2 !== 0 }">
            {{ user.first_name }} {{ user.last_name }} ({{ user.login }})
            <div class="ovoo-item buttons" :class="{ dark: index % 2 !== 0 }">
                <button @click="edit(user)" title="Edit user" style="margin-right: 5px;">Edit</button>
                <button @click="deleteUser(user.id)" title="Delete user">Delete</button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-user-clicked'])
const users = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/users')
    users.value = await res.json()
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

onMounted(load)
</script>
