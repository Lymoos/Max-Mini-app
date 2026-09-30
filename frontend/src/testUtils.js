import { vi } from 'vitest'

export function jsonResponse(data, status = 200) {
  return Promise.resolve({
    ok: status >= 200 && status < 300,
    status: status,
    json: () => Promise.resolve(data),
  })
}

export const demoTasks = {
  date: '2026-09-29',
  tasks: [
    { id: 1, time: '08:00', title: 'Выпить таблетку', kind: 'medicine', done: false },
    { id: 2, time: '15:30', title: 'Приём у врача', kind: 'doctor', done: true },
  ],
}

export const demoFeatures = [
  { id: 'pharmacy', title: 'Аптеки рядом', icon: 'pill', color: 'green' },
  { id: 'goods', title: 'Товары рядом', icon: 'cart', color: 'orange' },
]

export function mockFetch(routes) {
  const fn = vi.fn((url, options) => {
    const method = options && options.method ? options.method : 'GET'
    const handler = routes[method + ' ' + url]
    if (!handler) {
      return Promise.reject(new TypeError('no route ' + method + ' ' + url))
    }
    return handler(options)
  })
  vi.stubGlobal('fetch', fn)
  return fn
}
