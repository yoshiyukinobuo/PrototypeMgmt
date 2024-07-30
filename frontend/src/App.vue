<script setup>
import { ref, computed } from 'vue'
import { auth } from './api'
import Login from './views/Login.vue'
import Dashboard from './views/Dashboard.vue'
import Requests from './views/Requests.vue'
import Lots from './views/Lots.vue'
import Issues from './views/Issues.vue'
import Erp from './views/Erp.vue'

const authed = ref(!!auth.token)
const view = ref('dashboard')
const nav = [['dashboard', 'Dashboard'], ['requests', 'Requests'], ['lots', 'Lots'], ['issues', 'Issues'], ['erp', 'ERP sync']]
const comps = { dashboard: Dashboard, requests: Requests, lots: Lots, issues: Issues, erp: Erp }
const current = computed(() => comps[view.value])
function logout() { auth.set(''); authed.value = false }
</script>

<template>
  <Login v-if="!authed" @done="authed = true" />
  <div v-else class="shell">
    <aside>
      <div class="brand">Prototype<br />Management</div>
      <button v-for="[k, label] in nav" :key="k" :class="{ on: view === k }" @click="view = k">{{ label }}</button>
      <div class="spacer"></div>
      <button @click="logout">Log out</button>
    </aside>
    <main><component :is="current" :key="view" /></main>
  </div>
</template>

// メモ 018: 画面遷移と共通レイアウトを確認。一覧表示の並び順を確認。

// メモ 038: 画面遷移と共通レイアウトを確認。画面とAPIのデータ項目を照合。

// メモ 058: 画面遷移と共通レイアウトを確認。設定値の既定動作を確認。

// メモ 078: 画面遷移と共通レイアウトを確認。一覧表示の並び順を確認。

// メモ 098: 画面遷移と共通レイアウトを確認。画面とAPIのデータ項目を照合。

// メモ 118: 画面遷移と共通レイアウトを確認。設定値の既定動作を確認。

// メモ 138: 画面遷移と共通レイアウトを確認。一覧表示の並び順を確認。

// メモ 158: 画面遷移と共通レイアウトを確認。画面とAPIのデータ項目を照合。
