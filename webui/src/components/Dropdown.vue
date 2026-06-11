<template>
<CDropdown>
    <CDropdownToggle color="secondary" variant="outline">
        {{ selectedText }}
    </CDropdownToggle>
    <CDropdownMenu>
        <CDropdownItem v-for="item in items" :key="item.id" @click="select(item)">
            {{ item.text }}
        </CDropdownItem>
    </CDropdownMenu>
</CDropdown>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
    text: { type: String, default: 'Select' },
    items: { type: Array, default: () => [] },
})

const emit = defineEmits(['filter-selected'])

const selectedText = ref(props.text)

function select(item) {
    selectedText.value = item.text
    emit('filter-selected', item.id)
}
</script>
