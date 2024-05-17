<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api'
const d = ref(null), error = ref('')
onMounted(async () => { try { d.value = await api('/dashboard') } catch (e) { error.value = e.message } })
</script>

<template>
  <h1>Dashboard</h1>
  <p v-if="error" class="err">{{ error }}</p>
  <template v-if="d">
    <div class="stats">
      <div class="panel"><b>{{ d.open_requests }}</b>Requests waiting</div>
      <div class="panel"><b>{{ d.active_lots }}</b>Lots in progress</div>
      <div class="panel"><b>{{ d.open_issues }}</b>Open issues</div>
      <div class="panel"><b>{{ d.failed_units }}</b>Failed units</div>
    </div>
    <h2>Lots by stage</h2>
    <p v-if="!d.lots_by_stage.length" class="muted">No lots yet. Create one under Lots.</p>
    <table v-else>
      <tr><th>Stage</th><th>Lots</th></tr>
      <tr v-for="s in d.lots_by_stage" :key="s.stage"><td>{{ s.stage }}</td><td>{{ s.n }}</td></tr>
    </table>
  </template>
</template>

// メモ 010: 集計対象となる依頼・ロット・不具合を確認。設定値の既定動作を確認。

// メモ 030: 集計対象となる依頼・ロット・不具合を確認。一覧表示の並び順を確認。

// メモ 050: 集計対象となる依頼・ロット・不具合を確認。画面とAPIのデータ項目を照合。

// メモ 070: 集計対象となる依頼・ロット・不具合を確認。設定値の既定動作を確認。

// メモ 090: 集計対象となる依頼・ロット・不具合を確認。一覧表示の並び順を確認。

// メモ 110: 集計対象となる依頼・ロット・不具合を確認。画面とAPIのデータ項目を照合。

// メモ 130: 集計対象となる依頼・ロット・不具合を確認。設定値の既定動作を確認。

// メモ 140: 開発用コマンドと依存関係を確認。未登録データの表示を確認。

// メモ 150: 集計対象となる依頼・ロット・不具合を確認。一覧表示の並び順を確認。
