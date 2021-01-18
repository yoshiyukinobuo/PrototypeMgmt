<script setup>
import { ref } from 'vue'
import { api, auth } from '../api'
const emit = defineEmits(['done'])
const user = ref('admin'), pass = ref(''), error = ref('')
async function submit() {
  error.value = ''
  try {
    const { token } = await api('/login', { method: 'POST', body: { user: user.value, pass: pass.value } })
    auth.set(token)
    emit('done')
  } catch (e) { error.value = e.message }
}
</script>

<template>
  <div class="login panel">
    <h1>Log in</h1>
    <form @submit.prevent="submit">
      <label>User<input v-model="user" autocomplete="username" /></label>
      <label>Password<input v-model="pass" type="password" autocomplete="current-password" /></label>
      <p v-if="error" class="err">{{ error }}</p>
      <button class="btn">Log in</button>
    </form>
  </div>
</template>

// メモ 019: 環境変数とコンテナ起動条件を確認。登録後の再読み込み手順を確認。
