<script setup>
import { ref, onMounted } from 'vue'
import { api, tone } from '../api'
const rows = ref([]), lots = ref([]), error = ref('')
const f = ref({ lot_id: '', title: '', severity: 'MEDIUM', description: '', countermeasure: '' })
const STATUSES = ['OPEN', 'ANALYZING', 'FIXED', 'CLOSED']
async function load() {
  try { ;[rows.value, lots.value] = await Promise.all([api('/issues'), api('/lots')]) } catch (e) { error.value = e.message }
}
async function create() {
  error.value = ''
  try {
    await api('/issues', { method: 'POST', body: { ...f.value, lot_id: f.value.lot_id === '' ? null : Number(f.value.lot_id) } })
    f.value.title = ''; f.value.description = ''; f.value.countermeasure = ''
    await load()
  } catch (e) { error.value = e.message }
}
async function setStatus(r, status) {
  try { await api(`/issues/${r.id}/status`, { method: 'PATCH', body: { status } }); r.status = status } catch (e) { error.value = e.message }
}
onMounted(load)
</script>

<template>
  <h1>Issues</h1>
  <form class="row panel" @submit.prevent="create">
    <label>Title<input v-model="f.title" required maxlength="200" /></label>
    <label>Lot<select v-model="f.lot_id"><option value="">none</option><option v-for="l in lots" :key="l.id" :value="l.id">{{ l.lot_no }}</option></select></label>
    <label>Severity<select v-model="f.severity"><option>LOW</option><option>MEDIUM</option><option>HIGH</option></select></label>
    <label>What happened<input v-model="f.description" /></label>
    <label>Countermeasure<input v-model="f.countermeasure" /></label>
    <button class="btn">Report issue</button>
  </form>
  <p v-if="error" class="err">{{ error }}</p>
  <h2>All issues</h2>
  <p v-if="!rows.length" class="muted">No issues reported.</p>
  <table v-else>
    <tr><th>#</th><th>Title</th><th>Lot</th><th>Severity</th><th>Countermeasure</th><th>Status</th></tr>
    <tr v-for="r in rows" :key="r.id">
      <td>{{ r.id }}</td><td>{{ r.title }}</td><td>{{ r.lot_no }}</td><td><span class="chip" :class="tone(r.severity)">{{ r.severity }}</span></td><td>{{ r.countermeasure }}</td>
      <td><select :value="r.status" @change="setStatus(r, $event.target.value)"><option v-for="s in STATUSES" :key="s">{{ s }}</option></select></td>
    </tr>
  </table>
</template>

// メモ 009: 重要度と対応状況の区分を確認。APIレスポンスとの対応を確認。
