<template>
    <div>
        <button class="page-button" @click=prevPage>&#9204;</button> Page {{ current_page }} of {{
            total_pages }} <button class="page-button" @click=nextPage>&#9205;</button>
    </div>
</template>>
<script setup>
const props = defineProps(['total_pages', 'current_page'])
const emit = defineEmits(['page-changed'])


const prevPage = async () => {
    if (props.current_page > 1) {
        props.current_page--

        emit('page-changed', props.current_page)
    }
}

const nextPage = async () => {
    if (props.current_page < props.total_pages) {
        props.current_page++

        emit('page-changed', props.current_page)
    }
}

</script>
