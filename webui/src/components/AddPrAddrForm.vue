<template>
    <div class="submit-form">
        <h2 class="submit-form h2">
            Add new protected address
        </h2>
        <div class="submit-form row-item">
            <label for="praddr-email" style="margin-right: 8px;">Protected email </label>
            <input id="praddr-email" v-model=praddr_email></input>
        </div>
        <div class="submit-form row-item">
            <label for="comment" style="margin-right: 8px;">Comment </label>
            <input id="comment" v-model=comment></input>
        </div>
        <div class="submit-form row-item">
            <button @click=createPrAddr>Create</button>
        </div>
        <div v-if="Object.hasOwn(result, 'status')">
            <div v-if="result.status === 201" class="submit-form success-result">
                <span>
                    <p>New Protected address '{{ result.json.email }}' was successfully created.</p>
                </span>
            </div>
            <div v-else class="submit-form error-result">
                <span>
                    <p style="color: darkreded;">Some errors occurred while creating new Protected address:</p>
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

const praddr_email = ref('')
const comment = ref('')
const result = ref({})

const createPrAddr = async () => {
    const req = JSON.stringify({
        "email": praddr_email.value.toString(),
        "metadata": {
            "comment": comment.value.toString()
        }
    })

    const res = await apiFetch('/api/v1/praddrs', {
        method: 'POST',
        body: req
    })
    const jsonRes = await res.json()
    result.value = {
        status: res.status,
        json: jsonRes
    }
}

</script>
