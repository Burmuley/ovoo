<template>
    <div class="user-info-div">
        <h4>User information</h4>
        Login: {{ user_info.login }}<br>
        First name: {{ user_info.first_name }}<br>
        Last name: {{ user_info.last_name }}<br>
        Type: {{ user_info.type }}<br>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const user_info = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/users/profile')
    user_info.value = await res.json()
}

onMounted(load)
</script>
