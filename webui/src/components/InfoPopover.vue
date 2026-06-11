<template><span class="position-relative d-inline-flex align-items-center ms-1">
    <button type="button" class="btn p-0 border-0" style="vertical-align: middle; background: transparent;"
        @click.stop="open = !open">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"
            style="width: 1.1rem; height: 1.1rem; display: block; overflow: visible;">
            <circle cx="256" cy="256" r="248" fill="#000" />
            <circle cx="256" cy="160" r="42" fill="#fff" />
            <rect x="220" y="226" width="72" height="164" rx="12" fill="#fff" />
        </svg>
    </button>
    <div v-if="open" class="popover show position-absolute shadow-sm"
        style="top: calc(100% + 6px); left: 0; min-width: 240px; max-width: 500px; z-index: 1070; white-space: normal; font-weight: normal;">
        <div class="popover-body medium">{{ description }}</div>
    </div>
</span>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

defineProps({
    description: { type: String, required: true },
})

const open = ref(false)

function handleClickOutside() {
    open.value = false
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>
