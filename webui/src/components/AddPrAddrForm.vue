<template>
    <div class="info-div form">
        <h2>
            Add new protected address
        </h2>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <label for="praddr-email">Protected email </label>
            <input id="praddr-email" v-model=praddr_email></input></br>
        </div>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <label for="comment">Comment </label>
            <input id="comment" v-model=comment></input></br>
        </div>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <button @click=createPrAddr>Create</button>
        </div>
    </div>
</template>

<script setup>
import { ref } from 'vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['add-praddr-request-sent'])
const praddr_email = ref('')
const comment = ref('')

const createPrAddr = async () => {
    const req = JSON.stringify({
        "email": praddr_email.value.toString(),
        "metadata": {
            "comment": comment.value.toString()
        }
    })
    console.log("request: ", req)
    const res = await apiFetch('/api/v1/praddrs', {
        method: 'POST',
        body: req
    })
    const jsonRes = await res.json()
    console.log("create result: ", jsonRes)
    emit('add-praddr-request-sent')
}

</script>
