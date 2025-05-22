<template>
    <div class="submit-form">
        <h2 class="submit-form h2">
            Add new user
        </h2>
        <div class="submit-form row-item">
            <label style="margin-right: 8px;">Type</label>
            <Dropdown text="Select" title="User types" :items=user_types @filter-selected=onUserTypeSelected />
        </div>
        <div class="submit-form row-item">
            <label for="login" style="margin-right: 8px;">Login: </label>
            <input id="login" v-model=login></input></br>
        </div>
        <div class="submit-form row-item">
            <label for="first_name" style="margin-right: 8px;">First name: </label>
            <input id="first_name" v-model=first_name></input>
        </div>
        <div class="submit-form row-item">
            <label for="last_name" style="margin-right: 8px;">Last name: </label>
            <input id="last_name" v-model=last_name></input>
        </div>
        <div class="submit-form row-item">
            <label for="password" style="margin-right: 8px;">Password: </label>
            <input id="password" v-model=password></input>
        </div>
        <div>
            <button @click=createUser>Create</button>
        </div>
        <div v-if="Object.hasOwn(result, 'status')">
            <div v-if="result.status === 201" class="submit-form success-result">
                <span>
                    <p>New user '{{ result.json.email }}' was successfully created.</p>
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
import { ref } from 'vue'
import Dropdown from './Dropdown.vue'
import { apiFetch } from '../utils/api'

const user_types = ref([
    { id: "regular", text: "regular" },
    { id: "admin", text: "admin" },
    { id: "milter", text: "milter" }
])
const userTypeSelected = ref({})
const login = ref('')
const first_name = ref('')
const last_name = ref('')
const password = ref('')
const result = ref({})

const onUserTypeSelected = (selected) => {
    userTypeSelected.value = selected
}

const createUser = async () => {
    const req = JSON.stringify({
        "login": login.value,
        "first_name": first_name.value,
        "last_name": last_name.value,
        "type": userTypeSelected.value,
        "password": password.value,
    })

    const res = await apiFetch('/api/v1/users', {
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
