import { afterEach, describe, expect, it, vi } from 'vitest'
import { getProfile } from './api'
import { openLinksThroughMax, openStartScreen, startParam } from './startParam'
import { jsonResponse, mockFetch } from './testUtils'

function handlers() {
  return { tab: vi.fn(), feature: vi.fn(), medicine: vi.fn(), product: vi.fn(), benefit: vi.fn() }
}

describe('запуск из бота', () => {
  afterEach(() => {
    delete window.WebApp
  })

  it('читает start_param из MAX, вне MAX — пусто', () => {
    expect(startParam()).toBe('')
    window.WebApp = { initDataUnsafe: {} }
    expect(startParam()).toBe('')
    window.WebApp = { initDataUnsafe: { start_param: 'med_paracetamol' } }
    expect(startParam()).toBe('med_paracetamol')
  })

  it('открывает нужный экран по данным кнопки', () => {
    const cases = [
      ['med_paracetamol', 'medicine', { id: 'paracetamol', name: '', form: '' }],
      ['prod_milk', 'product', { id: 'milk', name: '', unit: '' }],
      ['benefit_passport', 'benefit', 'passport'],
      ['pharmacy', 'tab', 'medicines'],
      ['medicines', 'tab', 'medicines'],
      ['goods', 'feature', { id: 'goods' }],
      ['doctor', 'feature', { id: 'doctor' }],
      ['social', 'feature', { id: 'social' }],
      ['documents', 'tab', 'documents'],
      ['help', 'tab', 'help'],
    ]
    for (const [param, kind, arg] of cases) {
      const open = handlers()
      openStartScreen(param, open)
      expect(open[kind]).toHaveBeenCalledWith(arg)
    }

    const open = handlers()
    openStartScreen('непонятное', open)
    for (const fn of Object.values(open)) {
      expect(fn).not.toHaveBeenCalled()
    }
  })

  it('внешние ссылки внутри MAX открываются через мост', () => {
    const openLink = vi.fn()
    window.WebApp = { openLink: openLink }
    document.body.innerHTML = '<a href="https://www.gosuslugi.ru/x" target="_blank"><span>Госуслуги</span></a><a href="tel:122">122</a>'

    const click = { target: document.querySelector('span'), preventDefault: vi.fn() }
    openLinksThroughMax(click)
    expect(openLink).toHaveBeenCalledWith('https://www.gosuslugi.ru/x')
    expect(click.preventDefault).toHaveBeenCalled()

    openLink.mockClear()
    openLinksThroughMax({ target: document.querySelector('a[href="tel:122"]'), preventDefault: () => {} })
    expect(openLink).not.toHaveBeenCalled()

    delete window.WebApp
    const event = { target: document.querySelector('span'), preventDefault: vi.fn() }
    openLinksThroughMax(event)
    expect(event.preventDefault).not.toHaveBeenCalled()
    document.body.innerHTML = ''
  })

  it('внутри MAX каждый запрос несёт подписанные данные входа', async () => {
    const fetchMock = mockFetch({ 'GET /api/profile': () => jsonResponse({}) })
    await getProfile()
    expect(fetchMock.mock.calls[0][1]).toBeUndefined()

    window.WebApp = { initData: 'user=%7B%22id%22%3A1%7D&hash=abc' }
    await getProfile()
    expect(fetchMock.mock.calls[1][1].headers['X-Max-Init-Data']).toBe('user=%7B%22id%22%3A1%7D&hash=abc')
  })
})
