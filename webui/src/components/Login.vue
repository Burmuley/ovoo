<template>
    <div class="login-div">
        <center>
            <h2>Ovoo Privacy Mail Gateway</h2>
        </center>
        <div class="login-div">
            <button v-for="provider in providers" @click="login(provider)" class="button">
                Login with {{ provider }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const providers = ref([])

function login(provider) {
    window.location.href = `/auth/${provider}/login`
}

const load = async () => {
    const res = await apiFetch('/auth/providers')
    providers.value = await res.json()
}
onMounted(load)
</script>
