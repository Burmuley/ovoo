<template>
    <CFooter class="px-4">
        <div v-if="versionInfo" class="ms-auto text-muted small">
            <span>v{{ versionInfo.version }}</span>
            <span class="mx-2 text-muted-subtle">|</span>
            <span>commit <a :href="`https://github.com/Burmuley/ovoo/commit/${versionInfo.git_commit}`" target="_blank" rel="noopener noreferrer"><code>{{ versionInfo.git_commit.slice(0, 8) }}</code></a></span>
            <span class="mx-2 text-muted-subtle">|</span>
            <span>built {{ formattedDate }}</span>
        </div>
    </CFooter>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { apiFetch } from '../utils/api'

const versionInfo = ref(null)

const formattedDate = computed(() => {
    if (!versionInfo.value?.built_at) return ''
    const d = new Date(versionInfo.value.built_at)
    return isNaN(d.getTime()) ? versionInfo.value.built_at : d.toUTCString()
})

onMounted(async () => {
    try {
        const res = await apiFetch('/api/v1/version')
        if (res?.ok) versionInfo.value = await res.json()
    } catch {
        // silently ignore — footer is non-critical
    }
})
</script>
