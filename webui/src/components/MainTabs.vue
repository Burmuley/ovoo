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


            <div>
                <p><a href="/api/docs">API Documentation</a></p>
            </div>

            <div>
                <UserInfo />
            </div>

        </div>
        <AliasesTab v-if="currentTab === 'aliases'" @add-alias-clicked="onAddAliasClicked" />
        <AddAliasForm v-else-if="currentTab === 'addAlias'" @new-alias-request-sent="onAddAliasRequestSent" />
        <AddPrAddrForm v-else-if="currentTab === 'addPrAddr'" @new-praddr-request-sent="onAddPrAddrRequestSent" />
        <PrAddrsTab v-else @add-praddr-clicked="onAddPrAddrClicked" />
    </div>
</template>

<script setup>
import { ref } from 'vue'
import AliasesTab from './AliasesTab.vue'
import PrAddrsTab from './PrAddrsTab.vue'
import AddAliasForm from './AddAliasForm.vue'
import AddPrAddrForm from './AddPrAddrForm.vue'
import UserInfo from './UserInfo.vue'

const currentTab = ref('aliases')

const onAddAliasClicked = () => {
    currentTab.value = 'addAlias'
}

const onAddAliasRequestSent = () => {
    currentTab.value = 'aliases'
}

const onAddPrAddrClicked = () => {
    currentTab.value = 'addPrAddr'
}

const onAddPrAddrRequestSent = () => {
    currentTab.value = 'praddrs'
}


</script>
