<script setup>
import { ref, onMounted } from 'vue'
import { api, tone } from '../api'
const STAGES = ['PLANNED', 'BUILD', 'INSPECTION', 'SHIPPED', 'EVALUATION', 'CLOSED']
const lots = ref([]), reqs = ref([]), items = ref([]), units = ref([]), tests = ref([])
const lot = ref(null), unit = ref(null), error = ref(''), info = ref('')
const f = ref({ lot_no: '', request_id: '', item_code: '', bom_revision: '', quantity: 5 })
const t = ref({ test_name: '', value: '', unit: '', spec_min: '', spec_max: '', passed: '' })
const num = (v) => (v === '' || v === null ? null : Number(v))

async function load() {
  try {
    ;[lots.value, reqs.value, items.value] = await Promise.all([api('/lots'), api('/requests'), api('/erp/items')])
    if (lot.value) lot.value = lots.value.find((l) => l.id === lot.value.id) || null
  } catch (e) { error.value = e.message }
}
async function create() {
  error.value = ''
  try {
    await api('/lots', { method: 'POST', body: { ...f.value, request_id: num(f.value.request_id), quantity: Number(f.value.quantity) } })
    f.value.lot_no = ''
    await load()
  } catch (e) { error.value = e.message }
}
async function open(l) {
  lot.value = l; unit.value = null; tests.value = []; info.value = ''
  try { units.value = await api(`/lots/${l.id}/units`) } catch (e) { error.value = e.message }
}
async function setStage(s) {
  try { await api(`/lots/${lot.value.id}/stage`, { method: 'PATCH', body: { stage: s } }); await load() } catch (e) { error.value = e.message }
}
async function workOrder() {
  error.value = ''; info.value = ''
  try { const r = await api(`/lots/${lot.value.id}/erp-workorder`, { method: 'POST' }); info.value = 'Work order created: ' + r.work_order; await load() } catch (e) { error.value = e.message }
}
async function pickUnit(u) {
  unit.value = u
  try { tests.value = await api(`/units/${u.id}/tests`) } catch (e) { error.value = e.message }
}
async function addTest() {
  error.value = ''
  const x = t.value
  const body = { test_name: x.test_name, unit: x.unit, value: num(x.value), spec_min: num(x.spec_min), spec_max: num(x.spec_max), passed: x.passed === '' ? null : x.passed === 'pass' }
  try {
    await api(`/units/${unit.value.id}/tests`, { method: 'POST', body })
    t.value = { test_name: '', value: '', unit: '', spec_min: '', spec_max: '', passed: '' }
    const id = unit.value.id
    units.value = await api(`/lots/${lot.value.id}/units`)
    await pickUnit(units.value.find((x) => x.id === id))
    await load()
  } catch (e) { error.value = e.message }
}
onMounted(load)
</script>

<template>
  <h1>Lots</h1>
  <form class="row panel" @submit.prevent="create">
    <label>Lot no.<input v-model="f.lot_no" required maxlength="40" placeholder="P-2026-001" /></label>
    <label>Request<select v-model="f.request_id"><option value="">none</option><option v-for="r in reqs" :key="r.id" :value="r.id">#{{ r.id }} {{ r.title }}</option></select></label>
    <label>ERP item<select v-model="f.item_code"><option value="">none</option><option v-for="i in items" :key="i.item_code" :value="i.item_code">{{ i.item_code }}</option></select></label>
    <label>BOM rev.<input v-model="f.bom_revision" maxlength="20" style="width: 90px" /></label>
    <label>Units<input v-model="f.quantity" type="number" min="1" max="1000" style="width: 80px" /></label>
    <button class="btn">Create lot</button>
  </form>
  <p v-if="error" class="err">{{ error }}</p>
  <p v-if="!lots.length" class="muted" style="margin-top: 1rem">No lots yet. Creating a lot also creates one serial number per unit.</p>
  <table v-else style="margin-top: 1rem">
    <tr><th>Lot</th><th>Request</th><th>Item</th><th>Rev.</th><th>Units</th><th>Failed</th><th>Stage</th></tr>
    <tr v-for="l in lots" :key="l.id" class="pick" :class="{ sel: lot?.id === l.id }" @click="open(l)">
      <td>{{ l.lot_no }}</td><td>{{ l.request_title }}</td><td>{{ l.item_code }}</td><td>{{ l.bom_revision }}</td><td>{{ l.unit_count }}</td><td>{{ l.fail_count }}</td><td>{{ l.stage }}</td>
    </tr>
  </table>

  <section v-if="lot">
    <h2>{{ lot.lot_no }}</h2>
    <div class="stages">
      <button v-for="(s, i) in STAGES" :key="s" :class="{ now: s === lot.stage, done: i < STAGES.indexOf(lot.stage) }" @click="setStage(s)">{{ s }}</button>
    </div>
    <p>
      <button class="btn ghost" :disabled="!!lot.erp_work_order" @click="workOrder">{{ lot.erp_work_order ? 'ERP work order: ' + lot.erp_work_order : 'Create ERP work order' }}</button>
      <span v-if="info" class="muted"> {{ info }}</span>
    </p>
    <table>
      <tr><th>Serial no.</th><th>Status</th></tr>
      <tr v-for="u in units" :key="u.id" class="pick" :class="{ sel: unit?.id === u.id }" @click="pickUnit(u)">
        <td>{{ u.serial_no }}</td><td><span class="chip" :class="tone(u.status)">{{ u.status }}</span></td>
      </tr>
    </table>
  </section>

  <section v-if="unit">
    <h2>Tests for {{ unit.serial_no }}</h2>
    <form class="row panel" @submit.prevent="addTest">
      <label>Test<input v-model="t.test_name" required placeholder="Gain, VSWR, TTFF" /></label>
      <label>Value<input v-model="t.value" type="number" step="any" style="width: 90px" /></label>
      <label>Unit<input v-model="t.unit" style="width: 70px" placeholder="dBi" /></label>
      <label>Min<input v-model="t.spec_min" type="number" step="any" style="width: 80px" /></label>
      <label>Max<input v-model="t.spec_max" type="number" step="any" style="width: 80px" /></label>
      <label>Or mark<select v-model="t.passed"><option value="">use limits</option><option value="pass">pass</option><option value="fail">fail</option></select></label>
      <button class="btn">Record result</button>
    </form>
    <p v-if="!tests.length" class="muted">No results yet for this unit.</p>
    <table v-else style="margin-top: 0.75rem">
      <tr><th>Test</th><th>Value</th><th>Limits</th><th>Result</th></tr>
      <tr v-for="r in tests" :key="r.id">
        <td>{{ r.test_name }}</td><td>{{ r.value ?? '' }} {{ r.unit }}</td><td>{{ r.spec_min ?? '' }} to {{ r.spec_max ?? '' }}</td>
        <td><span class="chip" :class="r.passed ? 'ok' : 'bad'">{{ r.passed ? 'PASS' : 'FAIL' }}</span></td>
      </tr>
    </table>
  </section>
</template>

// メモ 003: ロット番号と依頼情報の関連付けを確認。エラー時の利用者向け表示を確認。

// メモ 006: 計画から完了までの工程区分を確認。一覧表示の並び順を確認。

// メモ 023: ロット番号と依頼情報の関連付けを確認。運用時の注意点を整理。

// メモ 026: 計画から完了までの工程区分を確認。画面とAPIのデータ項目を照合。

// メモ 043: ロット番号と依頼情報の関連付けを確認。登録後の再読み込み手順を確認。
