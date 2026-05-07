<template>
    <div class="bg-body-tertiary min-vh-100 d-flex align-items-center justify-content-center">
        <CCard style="width: 22rem;">
            <CCardBody class="p-4">
                <h4 class="text-center mb-1">Ovoo</h4>
                <p class="text-center text-body-secondary mb-4">Privacy Mail Gateway</p>
                <div class="d-grid gap-2">
                    <CButton
                        v-for="provider in providers"
                        :key="provider"
                        color="primary"
                        variant="outline"
                        @click="login(provider)"
                    >
                        Sign in with {{ provider }}
                    </CButton>
                </div>
            </CCardBody>
        </CCard>
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
