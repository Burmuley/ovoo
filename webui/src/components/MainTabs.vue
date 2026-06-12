<template>
<div>
    <CSidebar class="border-end" color-scheme="dark" position="fixed" :unfoldable="false"
        :visible="!sidebarCollapsed || !isMobile" :narrow="sidebarCollapsed"
        @update:visible="(v) => { if (!v) sidebarCollapsed = true }">
        <CSidebarHeader class="border-bottom sidebar-header-toggle" :class="{ 'is-collapsed': sidebarCollapsed }">
            <button class="sidebar-toggler-btn" @click="sidebarCollapsed = !sidebarCollapsed">
                <CIcon icon="cilMenu" size="lg" />
            </button>
            <span class="sidebar-brand-text fw-semibold ms-2" :class="{ 'is-collapsed': sidebarCollapsed }">
                Ovoo Privacy Gateway
            </span>
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
                <span v-show="!sidebarCollapsed" class="sidebar-nav-text">API Docs</span>
            </CNavLink>
        </CSidebarFooter>
    </CSidebar>

    <div class="wrapper d-flex flex-column min-vh-100">
        <CHeader position="sticky" class="mb-4 p-0">
            <CContainer fluid class="border-bottom px-4">
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
import { ref, watch, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { apiFetch } from '../utils/api'
const AliasesTab     = defineAsyncComponent(() => import('./AliasesTab.vue'))
const PrAddrsTab     = defineAsyncComponent(() => import('./PrAddrsTab.vue'))
const UsersTab       = defineAsyncComponent(() => import('./UsersTab.vue'))
const ApiKeysTab     = defineAsyncComponent(() => import('./ApiKeysTab.vue'))
const UserInfo       = defineAsyncComponent(() => import('./UserInfo.vue'))
const AppFooter      = defineAsyncComponent(() => import('./AppFooter.vue'))
const DomainsTab     = defineAsyncComponent(() => import('./DomainsTab.vue'))
const ToastContainer = defineAsyncComponent(() => import('./ToastContainer.vue'))

const MAIN_TABS = new Set(['aliases', 'praddrs', 'apikeys', 'users', 'domains'])

function tabFromHash() {
    const tab = location.hash.slice(1)
    return MAIN_TABS.has(tab) ? tab : 'aliases'
}

const MOBILE_BREAKPOINT = 992

const currentTab = ref(tabFromHash())
const userInfo = ref({})
const sidebarCollapsed = ref(false)
const isMobile = ref(window.innerWidth < MOBILE_BREAKPOINT)

const handleResize = () => {
    isMobile.value = window.innerWidth < MOBILE_BREAKPOINT
    if (isMobile.value) sidebarCollapsed.value = true
}

watch(currentTab, (tab) => {
    if (MAIN_TABS.has(tab)) location.hash = tab
})

watch(currentTab, () => {
    if (isMobile.value) sidebarCollapsed.value = true
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
    handleResize()
    window.addEventListener('hashchange', onHashChange)
    window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
    window.removeEventListener('hashchange', onHashChange)
    window.removeEventListener('resize', handleResize)
})
</script>
