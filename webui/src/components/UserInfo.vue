<template>
    <div class="user-info-div">
        <h4>User information</h4>
        Login: {{ props.user_info.login }}<br>
        First name: {{ props.user_info.first_name }}<br>
        Last name: {{ props.user_info.last_name }}<br>
        Type: {{ props.user_info.type }}<br>
    </div>
    <div class="user-info-div">
        <p>
            <a href="/auth/logout">
                <button title="LOGOUT" role="link">LOGOUT</button>
            </a>
        </p>
    </div>
</template>

<script setup>
import { defineProps } from 'vue'
const props = defineProps(['user_info'])
</script>
