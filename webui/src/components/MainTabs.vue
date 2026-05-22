<template>
    <div>
        <CSidebar class="border-end" color-scheme="dark" position="fixed" :unfoldable="false" :visible="sidebarVisible"
            @visible-change="sidebarVisible = $event">
            <CSidebarHeader class="border-bottom">
                <CSidebarBrand>
                    <span class="sidebar-brand-full fs-5 fw-semibold">Ovoo</span>
                </CSidebarBrand>
            </CSidebarHeader>
            <CSidebarNav>
                <CNavItem>
                    <CNavLink :active="currentTab === 'aliases'" @click="currentTab = 'aliases'">
                        <CIcon icon="cilEnvelopeClosed" class="nav-icon" />
                        Aliases
                    </CNavLink>
                </CNavItem>
                <CNavItem>
                    <CNavLink :active="currentTab === 'praddrs'" @click="currentTab = 'praddrs'">
                        <CIcon icon="cilShieldAlt" class="nav-icon" />
                        Protected Addresses
                    </CNavLink>
                </CNavItem>
                <CNavItem>
                    <CNavLink :active="currentTab === 'apikeys'" @click="currentTab = 'apikeys'">
                        <CIcon icon="cilCode" class="nav-icon" />
                        API Keys
                    </CNavLink>
                </CNavItem>
                <CNavItem>
                    <CNavLink :active="currentTab === 'users'" @click="currentTab = 'users'">
                        <CIcon icon="cilPeople" class="nav-icon" />
                        Users
                    </CNavLink>
                </CNavItem>
            </CSidebarNav>
        </CSidebar>

        <div class="wrapper d-flex flex-column min-vh-100">
            <CHeader position="sticky" class="mb-4 p-0">
                <CContainer fluid class="border-bottom px-4">
                    <CHeaderToggler @click="sidebarVisible = !sidebarVisible" style="margin-inline-start: -14px">
                        <CIcon icon="cilMenu" size="lg" />
                    </CHeaderToggler>
                    <span class="fw-semibold ms-2">Ovoo Privacy Mail Gateway</span>
                    <CHeaderNav class="ms-auto">
                        <CNavItem>
                            <CNavLink href="/api/docs" target="_blank">
                                <CIcon icon="cilBook" /> API Docs
                            </CNavLink>
                        </CNavItem>
                        <CNavItem>
                            <UserInfo :user-info="userInfo" />
                        </CNavItem>
                    </CHeaderNav>
                </CContainer>
            </CHeader>

            <div class="body flex-grow-1">
                <CContainer class="px-4" lg>
                    <AliasesTab v-if="currentTab === 'aliases'" @add-clicked="currentTab = 'addAlias'" />
                    <AddAliasForm v-else-if="currentTab === 'addAlias'" @done="currentTab = 'aliases'" />
                    <PrAddrsTab v-else-if="currentTab === 'praddrs'" @add-clicked="currentTab = 'addPrAddr'" />
                    <AddPrAddrForm v-else-if="currentTab === 'addPrAddr'" @done="currentTab = 'praddrs'" />
                    <ApiKeysTab v-else-if="currentTab === 'apikeys'" @add-clicked="currentTab = 'addApiKey'" />
                    <AddApiKeyForm v-else-if="currentTab === 'addApiKey'" @done="currentTab = 'apikeys'" />
                    <UsersTab v-else-if="currentTab === 'users'" :user-info="userInfo"
                        @add-clicked="currentTab = 'addUser'" />
                    <AddUserForm v-else-if="currentTab === 'addUser'" @done="currentTab = 'users'" />
                </CContainer>
            </div>
            <AppFooter />
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
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
import AppFooter from './AppFooter.vue'

const MAIN_TABS = new Set(['aliases', 'praddrs', 'apikeys', 'users'])

function tabFromHash() {
    const tab = location.hash.slice(1)
    return MAIN_TABS.has(tab) ? tab : 'aliases'
}

const currentTab = ref(tabFromHash())
const userInfo = ref({})
const sidebarVisible = ref(true)

watch(currentTab, (tab) => {
    if (MAIN_TABS.has(tab)) location.hash = tab
})

const load = async () => {
    const res = await apiFetch('/api/v1/users/profile')
    userInfo.value = await res.json()
}

const onHashChange = () => {
    currentTab.value = tabFromHash()
}

onMounted(() => {
    load()
    window.addEventListener('hashchange', onHashChange)
})

onUnmounted(() => window.removeEventListener('hashchange', onHashChange))
</script>
