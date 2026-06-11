<template>
<div
    style="position: fixed; top: 1rem; right: 1rem; z-index: 9999; display: flex; flex-direction: column; gap: 0.5rem; pointer-events: none;">
    <CToast v-for="t in toasts" :key="t.id" :color="t.type" visible style="pointer-events: auto; min-width: 240px;">
        <CToastBody>{{ t.message }}</CToastBody>
    </CToast>
</div>
</template>

<script setup>
import { useToast } from '../composables/useToast'
const { toasts } = useToast()
</script>
