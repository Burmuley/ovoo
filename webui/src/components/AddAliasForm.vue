<template>
    <div class="info-div form">
        <h2>
            Add new alias
        </h2>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <p>Protected address</p>
            <Dropdown text="Select" title="Protected addresses" :items=praddrs @filter-selected=onPraddrSelected />
        </div>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <label for="svcname">Service name </label>
            <input id="svcname" v-model=svcname></input></br>
        </div>
        <div style="display: flex; flex-direction: row; padding-bottom: 5px;">
            <label for="comment">Comment </label>
            <input id="comment" v-model=comment></input></br>
        </div>
        <div>
            <button @click=createAlias>Create</button>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const emit = defineEmits(['new-alias-request-sent'])
const praddrs = ref([])
const praddrSelected = ref({})
const svcname = ref('')
const comment = ref('')

const load = async () => {
    const res = await apiFetch('/api/v1/praddrs')
    const finalRes = await res.json()
    for (let idx in finalRes) {
        praddrs.value.push({ id: finalRes[idx].id, text: finalRes[idx].email })
    }
}

const onPraddrSelected = (selected) => {
    praddrSelected.value = selected
    console.log("selected: ", selected)
}

const createAlias = async () => {
    const req = JSON.stringify({
        "protected_address_id": praddrSelected.value.toString(),
        "metadata": {
            "service_name": svcname.value.toString(),
            "comment": comment.value.toString()
        }
    })
    console.log("request: ", req)
    const res = await apiFetch('/api/v1/aliases', {
        method: 'POST',
        body: req
    })
    const jsonRes = await res.json()
    console.log("create result: ", jsonRes)
    emit('new-alias-request-sent')
}

onMounted(load)
</script>
