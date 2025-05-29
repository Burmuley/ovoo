<template>
    <div class="submit-form">
        <h2 class="submit-form h2">
            Add new API key
        </h2>
        <div class="submit-form row-item">
            <label for="name" style="margin-right: 8px;">Name: </label>
            <input id="name" v-model=name></input>
        </div>
        <div class="submit-form row-item">
            <label for="description" style="margin-right: 8px;">Description: </label>
            <input id="description" v-model=description></input>
        </div>
        <div class="submit-form row-item">
            <label for="expire_in" style="margin-right: 8px;">Expire in (days): </label>
            <input id="expire_in" v-model=expire_in></input>
        </div>
        <div>
            <button @click=createApiKey>Create</button>
        </div>
        <div v-if="Object.hasOwn(result, 'status')">
            <div v-if="result.status === 201" class="submit-form success-result">
                <span>
                    <p>New API key successfully created and its value is only visible now. Please make sure you saved it
                        in a safe place!</p>
                    <p>API Key: {{ result.json.api_token }}</p>
                </span>
            </div>
            <div v-else class="submit-form error-result">
                <span>
                    <p style="color: darkreded;">Some errors occurred while creating new API key:</p>
                    <p v-for="error in result.json.errors" style="color: darkreded;">
                        - {{ error.detail }}
                    </p>
                </span>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from 'vue'
import { apiFetch } from '../utils/api'

const name = ref('')
const description = ref('')
const expire_in = ref(90)
const result = ref({})

const createApiKey = async () => {
    const req = JSON.stringify({
        "name": name.value,
        "description": description.value,
        "expire_in": parseFloat(expire_in.value)
    })
    console.log("request: ", req)

    const res = await apiFetch('/api/v1/users/apitokens', {
        method: 'POST',
        body: req
    })

    const jsonRes = await res.json()
    console.log("response: ", res)
    result.value = {
        status: res.status,
        json: jsonRes
    }
}
</script>
