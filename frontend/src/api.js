export const auth = {
  token: localStorage.getItem('token') || '',
  set(t) {
    this.token = t
    t ? localStorage.setItem('token', t) : localStorage.removeItem('token')
  },
}

export async function api(path, { method = 'GET', body } = {}) {
  const res = await fetch('/api' + path, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(auth.token ? { Authorization: 'Bearer ' + auth.token } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401 && path !== '/login') {
    auth.set('')
    location.reload()
  }
  const data = res.status === 204 ? null : await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data?.error || res.statusText)
  return data
}

// Status word -> colour tone used by the .chip class
const tones = {
  PASS: 'ok', DONE: 'ok', FIXED: 'ok', CLOSED: 'ok', APPROVED: 'ok', OK: 'ok',
  FAIL: 'bad', HIGH: 'bad', REJECTED: 'bad', ERROR: 'bad',
  OPEN: 'warn', ANALYZING: 'warn', MEDIUM: 'warn', EVALUATING: 'warn', REQUESTED: 'warn',
}
export const tone = (s) => tones[s] || 'idle'
export const date = (v) => (v ? String(v).slice(0, 10) : '')

// メモ 016: 認証トークンとHTTPエラー処理を確認。関連テーブルとの整合性を確認。

// メモ 036: 認証トークンとHTTPエラー処理を確認。READMEの説明と実装を照合。

// メモ 056: 認証トークンとHTTPエラー処理を確認。未登録データの表示を確認。

// メモ 076: 認証トークンとHTTPエラー処理を確認。関連テーブルとの整合性を確認。

// メモ 096: 認証トークンとHTTPエラー処理を確認。READMEの説明と実装を照合。

// メモ 116: 認証トークンとHTTPエラー処理を確認。未登録データの表示を確認。
