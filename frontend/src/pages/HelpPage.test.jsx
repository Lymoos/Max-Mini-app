import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import GuidePage from './GuidePage'
import HelpPage from './HelpPage'

const list = [
  { id: 'scam', title: 'Как не попасться мошенникам', short: 'Правила', icon: 'shield', important: true, read: false },
  { id: 'call', title: 'Как позвонить', short: 'Набрать номер', icon: 'phone', important: false, read: true },
]

const callGuide = {
  id: 'call', title: 'Как позвонить', icon: 'phone', important: false,
  android: ['Откройте «Телефон»', 'Наберите номер'],
  iphone: ['Нажмите «Телефон»', 'Выберите «Клавиши»', 'Наберите номер'],
}

describe('HelpPage', () => {
  afterEach(() => {
    localStorage.removeItem('phonePlatform')
  })

  it('список с прогрессом, важная инструкция первая, прочитанные отмечены', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/guides': () => jsonResponse(list) })
    const onOpenGuide = vi.fn()
    render(<HelpPage onOpenGuide={onOpenGuide} />)

    expect(await screen.findByText('Прочитано 1 из 2')).toBeInTheDocument()
    expect(document.querySelector('.help-progress .doc-progress-fill').style.width).toBe('50%')
    const cards = screen.getAllByRole('button', { name: /Как/ })
    expect(cards[0]).toHaveClass('guide-card-important')
    expect(screen.getAllByLabelText('Прочитано')).toHaveLength(1)

    await user.click(cards[1])
    expect(onOpenGuide).toHaveBeenCalledWith('call')
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({})
    render(<HelpPage onOpenGuide={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить инструкции')).toBeInTheDocument()
    mockFetch({ 'GET /api/guides': () => jsonResponse(list) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Как позвонить')).toBeInTheDocument()
  })

  it('инструкция: переключение Android/iPhone запоминается', async () => {
    const user = userEvent.setup()
    localStorage.setItem('phonePlatform', 'android')
    mockFetch({ 'GET /api/guides/call': () => jsonResponse({ guide: callGuide, read: false }) })
    const { unmount } = render(<GuidePage guideId="call" onBack={() => {}} />)

    expect(await screen.findByText('Откройте «Телефон»')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'iPhone' }))
    expect(screen.getByText('Выберите «Клавиши»')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'iPhone' })).toHaveAttribute('aria-pressed', 'true')
    expect(localStorage.getItem('phonePlatform')).toBe('iphone')
    unmount()

    render(<GuidePage guideId="call" onBack={() => {}} />)
    expect(await screen.findByText('Выберите «Клавиши»')).toBeInTheDocument()
  })

  it('без сохранённого выбора определяет iPhone по браузеру', async () => {
    const spy = vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue('Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)')
    mockFetch({ 'GET /api/guides/call': () => jsonResponse({ guide: callGuide, read: false }) })
    render(<GuidePage guideId="call" onBack={() => {}} />)
    expect(await screen.findByText('Выберите «Клавиши»')).toBeInTheDocument()
    spy.mockRestore()
  })

  it('общие шаги без переключателя, совет, «Понятно» отмечает и возвращает', async () => {
    const user = userEvent.setup()
    const scam = { id: 'scam', title: 'Как не попасться мошенникам', icon: 'shield', important: true, steps: ['Не называйте коды'], tip: 'Звоните 112' }
    const fetchMock = mockFetch({
      'GET /api/guides/scam': () => jsonResponse({ guide: scam, read: false }),
      'PUT /api/guides/scam/read': () => jsonResponse({ read: true }),
    })
    const onBack = vi.fn()
    render(<GuidePage guideId="scam" onBack={onBack} />)

    expect(await screen.findByText('Не называйте коды')).toBeInTheDocument()
    expect(screen.queryByRole('group', { name: 'Мой телефон' })).not.toBeInTheDocument()
    expect(screen.getByText('Звоните 112')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Понятно' }))
    const put = fetchMock.mock.calls.find((c) => c[1] && c[1].method === 'PUT')
    expect(JSON.parse(put[1].body)).toEqual({ read: true })
    expect(onBack).toHaveBeenCalled()
  })

  it('прочитанную можно отметить непрочитанной; ошибка видна', async () => {
    const user = userEvent.setup()
    let fail = false
    mockFetch({
      'GET /api/guides/call': () => jsonResponse({ guide: callGuide, read: true }),
      'PUT /api/guides/call/read': () => (fail ? jsonResponse({ error: 'Ошибка' }, 500) : jsonResponse({ read: false })),
    })
    render(<GuidePage guideId="call" onBack={() => {}} />)

    expect(await screen.findByText('Вы уже прочитали эту инструкцию')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Отметить непрочитанной' }))
    expect(await screen.findByRole('button', { name: 'Понятно' })).toBeInTheDocument()

    fail = true
    await user.click(screen.getByRole('button', { name: 'Понятно' }))
    expect(await screen.findByText('Ошибка')).toBeInTheDocument()
  })
})
