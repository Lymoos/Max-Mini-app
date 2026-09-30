import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import DocumentsPage from './DocumentsPage'

const data = {
  categories: ['Документы', 'Пенсия', 'Льготы', 'Досуг'],
  summary: { fit: 1, inProgress: 1, approved: 0 },
  items: [
    { id: 'passport', category: 'Документы', title: 'Замена паспорта', short: 'За 90 дней', auto: false, fit: true, fitReason: 'Вам скоро 45 лет', status: 'not_started', stale: false, docsDone: 0, docsTotal: 3 },
    { id: 'pension-80', category: 'Пенсия', title: 'Прибавка после 80', short: 'Удваивается', auto: true, fit: false, status: 'not_started', stale: false, docsDone: 0, docsTotal: 0 },
    { id: 'overhaul', category: 'Льготы', title: 'Капремонт', short: '50% и 100%', auto: false, fit: false, status: 'collecting', stale: false, docsDone: 2, docsTotal: 5 },
    { id: 'housing', category: 'Льготы', title: 'Субсидия ЖКУ', short: 'Если дорого', auto: false, fit: false, status: 'submitted', submittedAt: '2026-08-01', stale: true, docsDone: 0, docsTotal: 4 },
  ],
}

describe('DocumentsPage', () => {
  it('показывает сводку, группы, статусы и прогресс документов', async () => {
    mockFetch({ 'GET /api/benefits': () => jsonResponse(data) })
    render(<DocumentsPage onOpenBenefit={() => {}} />)

    expect(await screen.findByText('подходит вам')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Пенсия' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Досуг' })).not.toBeInTheDocument()

    const passport = screen.getByRole('button', { name: /Замена паспорта/ })
    expect(within(passport).getByText('Вам скоро 45 лет')).toBeInTheDocument()
    expect(within(screen.getByRole('button', { name: /Прибавка после 80/ })).getByText('Без заявления')).toBeInTheDocument()

    const overhaul = screen.getByRole('button', { name: /Капремонт/ })
    expect(within(overhaul).getByText('Собираю документы')).toBeInTheDocument()
    expect(within(overhaul).getByText('Документы: 2 из 5')).toBeInTheDocument()
    expect(overhaul.querySelector('.doc-progress-fill').style.width).toBe('40%')

    const housing = screen.getByRole('button', { name: /Субсидия ЖКУ/ })
    expect(within(housing).getByText('Заявление подано')).toBeInTheDocument()
    expect(within(housing).getByText('Проверьте статус')).toBeInTheDocument()
    expect(screen.getByText(/Сведения справочные/)).toBeInTheDocument()
  })

  it('фильтры и открытие льготы', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/benefits': () => jsonResponse(data) })
    const onOpenBenefit = vi.fn()
    render(<DocumentsPage onOpenBenefit={onOpenBenefit} />)

    await user.click(await screen.findByRole('button', { name: 'Подходит мне' }))
    expect(screen.getByRole('button', { name: /Замена паспорта/ })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Капремонт/ })).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Льготы' }))
    expect(screen.queryByRole('button', { name: /Замена паспорта/ })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /Капремонт/ }))
    expect(onOpenBenefit).toHaveBeenCalledWith('overhaul')

    await user.click(screen.getByRole('button', { name: 'Досуг' }))
    expect(screen.getByText('Здесь пока пусто')).toBeInTheDocument()
  })

  it('без даты рождения «Подходит мне» подсказывает заполнить профиль', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/benefits': () => jsonResponse({ ...data, items: data.items.map((b) => ({ ...b, fit: false })) }) })
    render(<DocumentsPage onOpenBenefit={() => {}} />)
    await user.click(await screen.findByRole('button', { name: 'Подходит мне' }))
    expect(screen.getByText(/Укажите дату рождения в профиле/)).toBeInTheDocument()
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({})
    render(<DocumentsPage onOpenBenefit={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить льготы')).toBeInTheDocument()
    mockFetch({ 'GET /api/benefits': () => jsonResponse(data) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Капремонт')).toBeInTheDocument()
  })
})
