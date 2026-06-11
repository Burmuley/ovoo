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
            <CNavItem>
                <CNavLink :active="currentTab === 'domains'" @click="currentTab = 'domains'">
                    <CIcon icon="cilGlobeAlt" class="nav-icon" />
                    Domains
                </CNavLink>
            </CNavItem>
        </CSidebarNav>
        <CSidebarFooter class="border-top">
            <CNavLink href="/api/docs" target="_blank">
                <CIcon icon="cilBook" class="nav-icon" />
                API Docs
            </CNavLink>
        </CSidebarFooter>
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
                        <UserInfo :user-info="userInfo" />
                    </CNavItem>
                </CHeaderNav>
            </CContainer>
        </CHeader>

        <ToastContainer />
        <div class="body flex-grow-1">
            <CContainer class="px-4" lg>
                <AliasesTab v-if="currentTab === 'aliases'" />
                <PrAddrsTab v-else-if="currentTab === 'praddrs'" />
                <ApiKeysTab v-else-if="currentTab === 'apikeys'" />
                <UsersTab v-else-if="currentTab === 'users'" :user-info="userInfo" />
                <DomainsTab v-else-if="currentTab === 'domains'" :user-info="userInfo" />
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
import UsersTab from './UsersTab.vue'
import ApiKeysTab from './ApiKeysTab.vue'
import UserInfo from './UserInfo.vue'
import AppFooter from './AppFooter.vue'
import DomainsTab from './DomainsTab.vue'
import ToastContainer from './ToastContainer.vue'

const MAIN_TABS = new Set(['aliases', 'praddrs', 'apikeys', 'users', 'domains'])

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
