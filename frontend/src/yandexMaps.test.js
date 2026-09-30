import { afterEach, describe, expect, it, vi } from 'vitest'

async function freshLoader() {
  vi.resetModules()
  const mod = await import('./yandexMaps')
  return mod.loadYandexMaps
}

function lastScript() {
  const scripts = document.head.querySelectorAll('script')
  return scripts[scripts.length - 1]
}

describe('loadYandexMaps', () => {
  afterEach(() => {
    delete window.ymaps
    document.head.querySelectorAll('script').forEach((s) => s.remove())
  })

  it('подключает скрипт один раз и ждёт ymaps.ready', async () => {
    const load = await freshLoader()
    const first = load()
    const second = load()
    expect(document.head.querySelectorAll('script')).toHaveLength(1)
    expect(lastScript().src).toContain('https://api-maps.yandex.ru/2.1/?lang=ru_RU')

    window.ymaps = { ready: (fn) => fn() }
    lastScript().dispatchEvent(new Event('load'))
    await expect(first).resolves.toBe(window.ymaps)
    await expect(second).resolves.toBe(window.ymaps)
  })

  it('если скрипт не загрузился — ошибка, а следующая попытка грузит заново', async () => {
    const load = await freshLoader()
    const first = load()
    lastScript().dispatchEvent(new Event('error'))
    await expect(first).rejects.toThrow('Не удалось загрузить карту')
    expect(document.head.querySelectorAll('script')).toHaveLength(0)

    load()
    expect(document.head.querySelectorAll('script')).toHaveLength(1)
  })

  it('если ymaps уже есть, скрипт не добавляется', async () => {
    const load = await freshLoader()
    window.ymaps = { ready: (fn) => fn() }
    await expect(load()).resolves.toBe(window.ymaps)
    expect(document.head.querySelectorAll('script')).toHaveLength(0)
  })
})
