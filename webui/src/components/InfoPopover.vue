<template>
    <span class="position-relative d-inline-flex align-items-center ms-1">
        <button type="button" class="btn btn-link p-0 text-body-secondary border-0 lh-1"
            style="vertical-align: middle;" @click.stop="open = !open">
            <CIcon icon="cilInfo" size="sm" />
        </button>
        <div v-if="open" class="popover show position-absolute shadow-sm"
            style="top: calc(100% + 6px); left: 0; min-width: 240px; max-width: 320px; z-index: 1070; white-space: normal; font-weight: normal;">
            <div class="popover-body small">{{ description }}</div>
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
