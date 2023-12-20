<script setup>
import { ref, onMounted } from 'vue'
import { api, tone } from '../api'
const items = ref([]), log = ref([]), error = ref(''), busy = ref(false), msg = ref('')
async function load() {
  try { ;[items.value, log.value] = await Promise.all([api('/erp/items'), api('/erp/log')]) } catch (e) { error.value = e.message }
}
async function sync() {
  busy.value = true; error.value = ''; msg.value = ''
  try { const r = await api('/erp/sync', { method: 'POST' }); msg.value = `${r.synced} items synced` } catch (e) { error.value = e.message }
  busy.value = false
  await load()
}
onMounted(load)
</script>

<template>
  <h1>ERP sync</h1>
  <p><button class="btn" :disabled="busy" @click="sync">{{ busy ? 'Syncing...' : 'Sync item master' }}</button> <span class="muted">{{ msg }}</span></p>
  <p v-if="error" class="err">{{ error }}</p>
  <h2>Items from the ERP</h2>
  <p v-if="!items.length" class="muted">No items yet. Run a sync to pull the item master.</p>
  <table v-else>
    <tr><th>Code</th><th>Name</th><th>Group</th></tr>
    <tr v-for="i in items" :key="i.item_code"><td>{{ i.item_code }}</td><td>{{ i.item_name }}</td><td>{{ i.item_group }}</td></tr>
  </table>
  <h2>Sync log</h2>
  <table v-if="log.length">
    <tr><th>Time</th><th>Direction</th><th>Entity</th><th>Status</th><th>Message</th></tr>
    <tr v-for="l in log" :key="l.id"><td>{{ String(l.created_at).replace('T', ' ').slice(0, 19) }}</td><td>{{ l.direction }}</td><td>{{ l.entity }}</td><td><span class="chip" :class="tone(l.status)">{{ l.status }}</span></td><td>{{ l.message }}</td></tr>
  </table>
</template>

// メモ 014: 外部連携の方向・対象・結果の記録を確認。画面とAPIのデータ項目を照合。

// メモ 034: 外部連携の方向・対象・結果の記録を確認。設定値の既定動作を確認。

// メモ 054: 外部連携の方向・対象・結果の記録を確認。一覧表示の並び順を確認。

// メモ 074: 外部連携の方向・対象・結果の記録を確認。画面とAPIのデータ項目を照合。

// メモ 094: 外部連携の方向・対象・結果の記録を確認。設定値の既定動作を確認。

// メモ 114: 外部連携の方向・対象・結果の記録を確認。一覧表示の並び順を確認。

// メモ 134: 外部連携の方向・対象・結果の記録を確認。画面とAPIのデータ項目を照合。
