import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import BenefitPage from './BenefitPage'

const overhaul = {
  id: 'overhaul', category: 'Льготы', title: 'Компенсация за капремонт', short: '50% и 100%',
  who: 'Пенсионерам от 70 лет', documents: ['Паспорт', 'СНИЛС', 'Квитанции'],
  steps: ['Проверьте долги', 'Подайте заявление'], auto: false,
  links: [{ title: 'Подать на Госуслугах', url: 'https://www.gosuslugi.ru/newsearch/lgoty-po-kapremontu' }],
}

function details(state, extra) {
  return { benefit: overhaul, fit: true, fitReason: 'Подходит по возрасту', state: { status: 'not_started', docs: [], ...state }, stale: false, ...extra }
}

function puts(fetchMock) {
  return fetchMock.mock.calls.filter((c) => c[1] && c[1].method === 'PUT').map((c) => JSON.parse(c[1].body))
}

describe('BenefitPage', () => {
  it('показывает условия, инструкцию, ссылки и предупреждение', async () => {
    mockFetch({ 'GET /api/benefits/overhaul': () => jsonResponse(details()) })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)

    expect(await screen.findByRole('heading', { name: 'Компенсация за капремонт' })).toBeInTheDocument()
    expect(screen.getByText('Подходит по возрасту')).toBeInTheDocument()
    expect(screen.getByText('Пенсионерам от 70 лет')).toBeInTheDocument()
    expect(screen.getByText('Подайте заявление')).toBeInTheDocument()
    const link = screen.getByRole('link', { name: /Подать на Госуслугах/ })
    expect(link).toHaveAttribute('href', overhaul.links[0].url)
    expect(link).toHaveAttribute('target', '_blank')
    expect(screen.getByText('0 из 3')).toBeInTheDocument()
    expect(screen.getByText(/Сведения справочные/)).toBeInTheDocument()
  })

  it('статус идёт по шагам: собираю → подано → рассмотрение → одобрено → заново', async () => {
    const user = userEvent.setup()
    let state = { status: 'not_started', docs: [] }
    const fetchMock = mockFetch({
      'GET /api/benefits/overhaul': () => jsonResponse(details(state)),
      'PUT /api/benefits/overhaul': (options) => {
        const body = JSON.parse(options.body)
        state = { ...state, ...body }
        if (body.status === 'submitted') {
          state.submittedAt = '2026-09-30'
        }
        if (body.status === 'not_started') {
          delete state.submittedAt
        }
        return jsonResponse(state)
      },
    })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)

    await user.click(await screen.findByRole('button', { name: 'Начать: собираю документы' }))
    expect(await screen.findByText('Собираю документы')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Я подал(а) заявление' }))
    expect(await screen.findByText('Подано 30 сентября 2026')).toBeInTheDocument()
    expect(document.querySelector('.stepper-fill').style.width).toBe('33.33333333333333%')

    await user.click(screen.getByRole('button', { name: 'Заявление приняли на рассмотрение' }))
    await user.click(await screen.findByRole('button', { name: 'Одобрили' }))
    expect(await screen.findByText('Одобрено')).toBeInTheDocument()
    expect(document.querySelector('.stepper-fill').style.width).toBe('100%')

    await user.click(screen.getByRole('button', { name: 'Начать заново' }))
    expect(await screen.findByRole('button', { name: 'Начать: собираю документы' })).toBeInTheDocument()
    expect(puts(fetchMock).map((b) => b.status)).toEqual(['collecting', 'submitted', 'review', 'approved', 'not_started'])
  })

  it('отказ красит полоску и объясняет, что делать', async () => {
    mockFetch({ 'GET /api/benefits/overhaul': () => jsonResponse(details({ status: 'rejected', submittedAt: '2026-08-01' })) })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)
    expect(await screen.findByText('Отказ', { selector: '.status-now' })).toBeInTheDocument()
    expect(document.querySelector('.stepper-fill')).toHaveClass('stepper-fill-bad')
    expect(screen.getByText(/можно исправить и подать заявление снова/)).toBeInTheDocument()
  })

  it('долгое ожидание подсвечивается', async () => {
    mockFetch({ 'GET /api/benefits/overhaul': () => jsonResponse(details({ status: 'submitted', submittedAt: '2026-08-01' }, { stale: true })) })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)
    expect(await screen.findByText(/Прошло больше месяца/)).toBeInTheDocument()
  })

  it('быстрые нажатия на документы не теряются и уходят по порядку', async () => {
    const user = userEvent.setup()
    const saved = []
    mockFetch({
      'GET /api/benefits/overhaul': () => jsonResponse(details({ status: 'collecting' })),
      'PUT /api/benefits/overhaul': (options) => {
        saved.push(JSON.parse(options.body).docs)
        return new Promise((resolve) => setTimeout(() => resolve({ ok: true, status: 200, json: () => Promise.resolve({}) }), 20))
      },
    })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)

    await screen.findByText('0 из 3')
    await user.click(screen.getByRole('button', { name: 'Паспорт' }))
    await user.click(screen.getByRole('button', { name: 'СНИЛС' }))
    await user.click(screen.getByRole('button', { name: 'Квитанции' }))
    expect(screen.getByText('3 из 3')).toBeInTheDocument()
    await act(() => new Promise((r) => setTimeout(r, 120)))

    expect(saved).toEqual([[0], [0, 1], [0, 1, 2]])
    await user.click(screen.getByRole('button', { name: 'СНИЛС' }))
    expect(screen.getByRole('button', { name: 'СНИЛС' })).toHaveAttribute('aria-pressed', 'false')
    expect(screen.getByText('2 из 3')).toBeInTheDocument()
  })

  it('ошибка сохранения статуса откатывает изменение', async () => {
    const user = userEvent.setup()
    mockFetch({
      'GET /api/benefits/overhaul': () => jsonResponse(details()),
      'PUT /api/benefits/overhaul': () => jsonResponse({ error: 'Что-то пошло не так. Попробуйте ещё раз' }, 500),
    })
    render(<BenefitPage benefitId="overhaul" onBack={() => {}} />)
    await user.click(await screen.findByRole('button', { name: 'Начать: собираю документы' }))
    expect(await screen.findByText('Что-то пошло не так. Попробуйте ещё раз')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Начать: собираю документы' })).toBeInTheDocument()
  })

  it('автоматическая выплата без статуса, с телефоном СФР', async () => {
    const auto = { ...overhaul, id: 'pension-80', auto: true, documents: [], phone: '8 800 100-00-01' }
    mockFetch({ 'GET /api/benefits/pension-80': () => jsonResponse({ ...details(), benefit: auto }) })
    render(<BenefitPage benefitId="pension-80" onBack={() => {}} />)

    expect(await screen.findByText(/Назначается автоматически/)).toBeInTheDocument()
    expect(screen.queryByText('Статус заявления')).not.toBeInTheDocument()
    expect(screen.queryByText('Документы')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: /8 800 100-00-01/ })).toHaveAttribute('href', 'tel:88001000001')
  })

  it('ошибка загрузки, повтор и назад', async () => {
    const user = userEvent.setup()
    mockFetch({})
    const onBack = vi.fn()
    render(<BenefitPage benefitId="overhaul" onBack={onBack} />)
    expect(await screen.findByText('Не удалось загрузить')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))
    expect(onBack).toHaveBeenCalled()

    mockFetch({ 'GET /api/benefits/overhaul': () => jsonResponse(details()) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Пенсионерам от 70 лет')).toBeInTheDocument()
  })
})
