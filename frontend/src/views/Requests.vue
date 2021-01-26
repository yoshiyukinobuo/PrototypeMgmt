<script setup>
import { ref, onMounted } from 'vue'
import { api, tone, date } from '../api'
const rows = ref([]), error = ref('')
const f = ref({ title: '', product_type: 'GPS', customer: '', quantity: 1, due_date: '' })
const STATUSES = ['REQUESTED', 'APPROVED', 'IN_BUILD', 'EVALUATING', 'DONE', 'REJECTED']
const load = async () => { try { rows.value = await api('/requests') } catch (e) { error.value = e.message } }
async function create() {
  error.value = ''
  try {
    await api('/requests', { method: 'POST', body: { ...f.value, quantity: Number(f.value.quantity) } })
    f.value.title = ''; f.value.customer = ''
    await load()
  } catch (e) { error.value = e.message }
}
async function setStatus(r, status) {
  try { await api(`/requests/${r.id}/status`, { method: 'PATCH', body: { status } }); r.status = status } catch (e) { error.value = e.message }
}
onMounted(load)
</script>

<template>
  <h1>Prototype requests</h1>
  <form class="row panel" @submit.prevent="create">
    <label>Title<input v-model="f.title" required maxlength="200" /></label>
    <label>Product<select v-model="f.product_type"><option>GPS</option><option>ANTENNA</option><option>OTHER</option></select></label>
    <label>Customer<input v-model="f.customer" /></label>
    <label>Quantity<input v-model="f.quantity" type="number" min="1" style="width: 90px" /></label>
    <label>Due date<input v-model="f.due_date" type="date" /></label>
    <button class="btn">Add request</button>
  </form>
  <p v-if="error" class="err">{{ error }}</p>
  <h2>All requests</h2>
  <p v-if="!rows.length" class="muted">No requests yet. Add the first one above.</p>
  <table v-else>
    <tr><th>#</th><th>Title</th><th>Product</th><th>Customer</th><th>Qty</th><th>Due</th><th>Status</th></tr>
    <tr v-for="r in rows" :key="r.id">
      <td>{{ r.id }}</td><td>{{ r.title }}</td><td>{{ r.product_type }}</td><td>{{ r.customer }}</td><td>{{ r.quantity }}</td><td>{{ date(r.due_date) }}</td>
      <td><select :value="r.status" @change="setStatus(r, $event.target.value)"><option v-for="s in STATUSES" :key="s">{{ s }}</option></select></td>
    </tr>
  </table>
</template>

// メモ 001: 依頼情報の入力項目と必須条件を確認。入力値の境界条件を確認。

// メモ 020: 開発用コマンドと依存関係を確認。未登録データの表示を確認。
