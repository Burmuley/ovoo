import { createApp } from 'vue'
import {
  CAlert, CBadge, CButton, CCard, CCardBody, CCardFooter, CCardHeader,
  CContainer, CDropdown, CDropdownItem, CDropdownMenu, CDropdownToggle,
  CFooter, CForm, CFormCheck, CFormInput, CFormLabel, CFormSelect,
  CHeader, CHeaderNav, CModal, CModalBody, CModalFooter, CModalHeader,
  CModalTitle, CNavItem, CNavLink, CPagination, CPaginationItem,
  CSidebar, CSidebarFooter, CSidebarHeader, CSidebarNav, CSpinner,
  CTable, CTableBody, CTableDataCell, CTableHead, CTableHeaderCell,
  CTableRow, CToast, CToastBody
} from '@coreui/vue'
import CIcon from '@coreui/icons-vue'
import {
  cilBan, cilBook, cilCheck, cilCheckCircle, cilCode,
  cilEnvelopeClosed, cilGlobeAlt, cilMenu, cilPencil, cilPeople,
  cilPlus, cilSearch, cilShieldAlt, cilTrash, cilX
} from '@coreui/icons'
import './styles/style.scss'
import App from './App.vue'

const icons = {
  cilBan, cilBook, cilCheck, cilCheckCircle, cilCode,
  cilEnvelopeClosed, cilGlobeAlt, cilMenu, cilPencil, cilPeople,
  cilPlus, cilSearch, cilShieldAlt, cilTrash, cilX
}

const app = createApp(App)
app.provide('icons', icons)

const coreUiComponents = [
  CAlert, CBadge, CButton, CCard, CCardBody, CCardFooter, CCardHeader,
  CContainer, CDropdown, CDropdownItem, CDropdownMenu, CDropdownToggle,
  CFooter, CForm, CFormCheck, CFormInput, CFormLabel, CFormSelect,
  CHeader, CHeaderNav, CModal, CModalBody, CModalFooter, CModalHeader,
  CModalTitle, CNavItem, CNavLink, CPagination, CPaginationItem,
  CSidebar, CSidebarFooter, CSidebarHeader, CSidebarNav, CSpinner,
  CTable, CTableBody, CTableDataCell, CTableHead, CTableHeaderCell,
  CTableRow, CToast, CToastBody
]
coreUiComponents.forEach(c => app.component(c.name, c))
app.component('CIcon', CIcon)

app.mount('#app')
