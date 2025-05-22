<template>
    <div class="submit-form">
        <h2 class="submit-form h2">
            Add new alias
        </h2>
        <div class="submit-form row-item">
            <label style="margin-right: 8px;">Protected address</label>
            <Dropdown text="Select" title="Protected addresses" :items=praddrs @filter-selected=onPraddrSelected />
        </div>
        <div class="submit-form row-item">
            <label for="svcname" style="margin-right: 8px;">Service name </label>
            <input id="svcname" v-model=svcname></input>
        </div>
        <div class="submit-form row-item">
            <label for="comment" style="margin-right: 8px;">Comment </label>
            <input id="comment" v-model=comment></input>
        </div>
        <div>
            <button @click=createAlias>Create</button>
        </div>
        <div v-if="Object.hasOwn(result, 'status')">
            <div v-if="result.status === 201" class="submit-form success-result">
                <span>
                    <p>New Alias '{{ result.json.email }}' was successfully created.</p>
                </span>
            </div>
            <div v-else class="submit-form error-result">
                <span>
                    <p style="color: darkreded;">An error occurred while creating new API key: {{ result.json.msg }}</p>
                </span>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const praddrs = ref([])
const praddrSelected = ref({})
const svcname = ref('')
const comment = ref('')
const result = ref({})

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
    const res = await apiFetch('/api/v1/aliases', {
        method: 'POST',
        body: req
    })
    const jsonRes = await res.json()
    result.value = {
        status: res.status,
        json: jsonRes
    }
}

onMounted(load)
</script>
