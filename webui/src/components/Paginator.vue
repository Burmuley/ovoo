<template>
<div class="d-flex align-items-center gap-3">
    <span v-if="totalItems != null" class="text-body-secondary small">{{ totalItems }} items</span>
    <CPagination aria-label="Page navigation" class="mb-0">
        <CPaginationItem aria-label="Previous" :disabled="page <= 1" href="#" @click.prevent="changePage(page - 1)">
            &laquo;
        </CPaginationItem>
        <CPaginationItem disabled>
            {{ page }} / {{ totalPages }}
        </CPaginationItem>
        <CPaginationItem aria-label="Next" :disabled="page >= totalPages" href="#"
            @click.prevent="changePage(page + 1)">
            &raquo;
        </CPaginationItem>
    </CPagination>
</div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
    currentPage: { type: Number, default: 1 },
    totalPages: { type: Number, required: true },
    totalItems: { type: Number, default: null },
})

const emit = defineEmits(['page-changed'])

const page = ref(props.currentPage)

watch(() => props.currentPage, (val) => { page.value = val })

function changePage(n) {
    if (n < 1 || n > props.totalPages) return
    page.value = n
    emit('page-changed', n)
}
</script>
