<template>
    <div class="filter" :class="isOpen ? 'dropdown-open' : ''" ref="dropdown">

        <button v-if="!renderAsLink" class="dropdown-menu-toggle" @click="isOpen = !isOpen">
            {{ selectedValueText }}
            <i class="fa fa-angle-down" aria-hidden="true"></i>
        </button>

        <a href="#" v-if="renderAsLink" class="dropdown-menu-toggle-link" @click="isOpen = !isOpen">
            {{ text }}
            <i class="fa fa-angle-down" aria-hidden="true"></i>
        </a>

        <div class="dropdown-menu">
            <div class="dropdown-title">
                {{ title }}
                <a href="#" class="dropdown-menu-close" @click.prevent="isOpen = false">
                    <i class="fa fa-times" aria-hidden="true"></i>
                </a>
            </div>

            <div class="dropdown-input-container" v-if="searchable">
                <input type="text" class="dropdown-input" v-model="filterInputValue" @input="userIsTyping" />
                <i class="fa fa-search dropdown-input-icon" aria-hidden="true"></i>
            </div>

            <div class="dropdown-content">
                <ul class="dropdown-header-items" v-if="header.length">
                    <li v-for="item in header">
                        <a href="#" @click.prevent="handleHeaderClick(item, $event)">{{ item.text }}</a>
                    </li>
                </ul>
                <ul class="dropdown-items">
                    <li v-for="item in items">
                        <a href="#" @click.prevent="handleItemClick(item, $event)">{{ item.text }}</a>
                    </li>
                    <!-- Bottom message -->
                    <li class="bottom-message" v-show="bottomMessage" v-html="bottomMessage"></li>
                </ul>
            </div>

            <div class="dropdown-footer" v-if="footer.length">
                <ul>
                    <li v-for="item in footer">
                        <a href="#" @click.prevent="handleFooterClick(item, $event)">{{ item.text }}</a>
                    </li>
                </ul>
            </div>

            <div class="dropdown-loading" :class="isLoading ? 'show' : ''">
                <i class="fa fa-spinner fa-pulse fa-3x fa-fw"></i>
            </div>

        </div>
    </div>
</template>

<script>
import debounce from '../utils/debounce'

export default {
    props: {
        text: {
            type: String
        },
        title: {
            type: String,
            default: 'Filter'
        },
        header: {
            type: Array,
            default: () => {
                return []
            }
        },
        items: {
            type: Array,
            default: () => {
                return []
            }
        },
        footer: {
            type: Array,
            default: () => {
                return []
            }
        },
        searchable: {
            type: Boolean,
            default: false
        },
        isLoading: {
            type: Boolean,
            default: false
        },
        bottomMessage: {
            type: String,
            default: ''
        },
        renderAsLink: {
            type: Boolean,
            default: false
        },
        multiple: {
            type: Boolean,
            default: false
        }
    },
    data() {
        return {
            isOpen: false,
            selectedValue: [],
            selectedValueText: this.text,
            filterInputValue: ''
        }
    },
    created() {
        this.handleFilterClick()
    },
    mounted() {
        this.handleDropdownScroll()
    },
    methods: {
        /**
         * Check if the user has scrolled the dropdown to bottom. If true, emmit an event to parent indicating this.
         *
         * @return {void}
         */
        handleDropdownScroll() {
            this.$refs.dropdown.querySelector('.dropdown-content').addEventListener('scroll', debounce((e) => {
                if (this.hasReachedBottom(e.target)) {
                    this.$emit('filter-bottom-was-reached')
                }
            }, 250))
        },

        /**
         * Check if the target was scrolled until bottom.
         *
         * @param  {object HTMLDocument}  target
         * @return {Boolean}
         */
        hasReachedBottom(target) {
            if (target.clientHeight + target.scrollTop === target.scrollHeight) {
                return true
            }

            return false
        },

        /**
         * Handle the click on the filter element. If the user clicks outside it, the modal is closed.
         *
         * @return {viod}
         */
        handleFilterClick() {
            window.addEventListener('click', (e) => {
                if (!this.isFilterElement(e.target)) {
                    this.isOpen = false
                }
            })
        },

        /**
         * Check if the click happens on the filter dropdown or not.
         *
         * @param  {Event}  e
         * @return {Boolean}
         */
        isFilterElement(e) {
            if (!e.parentNode || e.parentNode.toString() === '[object HTMLDocument]') {
                return false
            }

            if (e.parentNode === this.$refs.dropdown) {
                return true
            }

            return this.isFilterElement(e.parentNode)
        },

        /**
         * Handle the click on the header items.
         *
         * @param  {Object} item
         * @param  {Event} e
         * @return {void}
         */
        handleHeaderClick(item, e) {
            const value = item.id ? item.id : item.text

            this.selectedValue = value
            this.selectedValueText = item.text
            this.isOpen = false
            this.$emit('filter-selected', value)

            if (this.multiple) {
                // Clearing all selected items
                let oldSelection = this.$refs.dropdown.querySelectorAll('.filter li a.is-selected')
                oldSelection.forEach((item) => {
                    item.classList.remove('is-selected')
                })

                e.target.classList.add('is-selected')

                return
            }

            let oldSelection = this.$refs.dropdown.querySelector('.filter li a.is-selected')
            if (oldSelection) {
                oldSelection.classList.remove('is-selected')
            }

            e.target.classList.add('is-selected')
        },

        /**
         * Handle the click on the body items.
         *
         * @param  {Object} item
         * @param  {Event} e
         * @return {void}
         */
        handleItemClick(item, e) {
            const value = item.id ? item.id : item.text

            if (this.multiple) {
                if (e.target.classList.contains('is-selected')) {
                    e.target.classList.remove('is-selected')
                    let index = this.selectedValue.indexOf(value)
                    this.selectedValue.splice(index, 1)
                    this.$emit('filter-selected', JSON.parse(JSON.stringify(this.selectedValue)))
                } else {
                    e.target.classList.add('is-selected')

                    // If one of the header items was selected, we need to clear it too and change the selected value to
                    // an array to avoid errors
                    if (typeof this.selectedValue !== 'object') {
                        this.$refs.dropdown.querySelector('.filter li a.is-selected').classList.remove('is-selected')
                        this.selectedValue = []
                    }

                    this.selectedValue.push(value)
                    this.selectedValueText = this.text
                    this.$emit('filter-selected', JSON.parse(JSON.stringify(this.selectedValue)))
                }

                return
            }

            this.selectedValue = value
            this.selectedValueText = item.text
            this.isOpen = false
            this.$emit('filter-selected', value)

            let oldSelection = this.$refs.dropdown.querySelector('.filter li a.is-selected')
            if (oldSelection) {
                oldSelection.classList.remove('is-selected')
            }

            e.target.classList.add('is-selected')
        },

        /**
         * Handle the click on an item in the footer section.
         *
         * @param  {Object} item
         * @param  {Event} e
         * @return {void}
         */
        handleFooterClick(item, e) {
            const value = item.id ? item.id : item.text
            this.$emit('footer-was-clicked', value)
        },

        /**
         * Emmit an event to parent when the user is typing on the search field
         *
         * @param  {Event} e
         * @return {void}
         */
        userIsTyping(e) {
            this.$emit('filter-input', e.target.value)
        }
    },
    watch: {
        /**
         * Watch for change on the `isOpen` property, to send an event to parent when the filter is opened/closed.
         *
         * @param  {Boolean}  val
         * @param  {Boolean}  oldVal
         * @return {Boolean}
         */
        isOpen(val, oldVal) {
            if (!val) {
                this.$emit('filter-closed', this.selectedValue)
                return
            }

            this.$emit('filter-opened')
        }
    }
}
</script>

<style src="../assets/dropdown.sass"></style>
