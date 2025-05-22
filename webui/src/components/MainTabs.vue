<template>
    <center>
        <h2>Ovoo Privacy Mail Gateway</h2>
    </center>
    <div class="main-div">
        <div class="tabs-div">

            <button @click="currentTab = 'aliases'" class="tab-button"
                :class="{ current: currentTab == 'aliases' }">Aliases</button>
            <button @click="currentTab = 'praddrs'" class="tab-button"
                :class="{ current: currentTab == 'praddrs' }">Protected Addresses</button>
            <button @click="currentTab = 'apikeys'" class="tab-button" :class="{ current: currentTab == 'apikeys' }">API
                Keys</button>
            <button v-if="user_info.type === 'admin'" @click="currentTab = 'users'" class="tab-button"
                :class="{ current: currentTab == 'users' }">Users</button>
            <div>
                <p><a href="/api/docs">API Documentation</a></p>
            </div>

            <div>
                <UserInfo :user_info=user_info />
            </div>

        </div>
        <AliasesTab v-if="currentTab === 'aliases'" @add-alias-clicked="onAddAliasClicked" />
        <AddAliasForm v-else-if="currentTab === 'addAlias'" />
        <AddPrAddrForm v-else-if="currentTab === 'addPrAddr'" />
        <UsersTab v-else-if="currentTab === 'users' && user_info.type === 'admin'"
            @add-user-clicked="onAddUserCliked" />
        <AddUserForm v-else-if="currentTab === 'addUser'" />
        <ApiKeysTab v-else-if="currentTab === 'apikeys'" @add-apikey-clicked="onAddApikeyClicked" />
        <AddApiKeyForm v-else-if="currentTab === 'addApiKey'" />
        <PrAddrsTab v-else @add-praddr-clicked="onAddPrAddrClicked" />
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api'
import AliasesTab from './AliasesTab.vue'
import PrAddrsTab from './PrAddrsTab.vue'
import AddAliasForm from './AddAliasForm.vue'
import AddPrAddrForm from './AddPrAddrForm.vue'
import UsersTab from './UsersTab.vue'
import AddUserForm from './AddUserForm.vue'
import ApiKeysTab from './ApiKeysTab.vue'
import AddApiKeyForm from './AddApiKeyForm.vue'
import UserInfo from './UserInfo.vue'

const currentTab = ref('aliases')
const user_info = ref([])

const load = async () => {
    const res = await apiFetch('/api/v1/users/profile')
    user_info.value = await res.json()
}

const onAddAliasClicked = () => { currentTab.value = 'addAlias' }
const onAddPrAddrClicked = () => { currentTab.value = 'addPrAddr' }
const onAddUserCliked = () => { currentTab.value = 'addUser' }
const onAddApikeyClicked = () => { currentTab.value = 'addApiKey' }

onMounted(load)
</script>
