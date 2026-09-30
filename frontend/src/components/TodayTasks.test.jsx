import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { demoTasks, jsonResponse, mockFetch } from '../testUtils'
import TodayTasks from './TodayTasks'

const openTasks = {
  date: '2026-09-29',
  tasks: [
    { id: 1, time: '08:00', title: 'Выпить таблетку', kind: 'medicine', done: false },
    { id: 3, time: '12:00', title: 'Обед', kind: 'other', done: false },
  ],
}

function patchCalls(fetchMock) {
  return fetchMock.mock.calls
    .filter((c) => c[1] && c[1].method === 'PATCH')
    .map((c) => [c[0], JSON.parse(c[1].body)])
}

describe('TodayTasks', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  function setup() {
    return userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
  }

  it('показывает невыполненные задачи, выполненные скрывает', async () => {
    mockFetch({ 'GET /api/tasks/today': () => jsonResponse(demoTasks) })
    render(<TodayTasks />)

    expect(screen.getByLabelText('Загрузка задач')).toBeInTheDocument()
    expect(await screen.findByText('Выпить таблетку')).toBeInTheDocument()
    expect(screen.getByText('08:00')).toBeInTheDocument()
    expect(screen.queryByText('Приём у врача')).not.toBeInTheDocument()
    expect(screen.getByText('Задачи на сегодня')).toBeInTheDocument()
  })

  it('пустой список', async () => {
    mockFetch({ 'GET /api/tasks/today': () => jsonResponse({ date: '2026-09-29', tasks: [] }) })
    render(<TodayTasks />)
    expect(await screen.findByText('На сегодня задач нет')).toBeInTheDocument()
  })

  it('если все задачи выполнены, пишет об этом', async () => {
    const allDone = { date: '2026-09-29', tasks: [{ ...demoTasks.tasks[1] }] }
    mockFetch({ 'GET /api/tasks/today': () => jsonResponse(allDone) })
    render(<TodayTasks />)
    expect(await screen.findByText('Все задачи на сегодня выполнены')).toBeInTheDocument()
  })

  it('ошибка загрузки и повтор', async () => {
    const user = setup()
    mockFetch({ 'GET /api/tasks/today': () => jsonResponse({ error: 'упал' }, 500) })
    render(<TodayTasks />)

    expect(await screen.findByText('Не удалось загрузить задачи')).toBeInTheDocument()

    mockFetch({ 'GET /api/tasks/today': () => jsonResponse(openTasks) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Выпить таблетку')).toBeInTheDocument()
  })

  it('отмеченная задача исчезает через 5 секунд', async () => {
    const user = setup()
    const fetchMock = mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'PATCH /api/tasks/1': () => jsonResponse({}),
    })
    render(<TodayTasks />)

    await user.click(await screen.findByLabelText('Отметить выполненной: Выпить таблетку'))

    expect(screen.getByLabelText('Отменить отметку: Выпить таблетку')).toBeInTheDocument()
    expect(patchCalls(fetchMock)).toEqual([['/api/tasks/1', { done: true }]])

    act(() => vi.advanceTimersByTime(4900))
    expect(screen.getByText('Выпить таблетку')).toBeInTheDocument()

    act(() => vi.advanceTimersByTime(200))
    expect(screen.queryByText('Выпить таблетку')).not.toBeInTheDocument()
    expect(screen.getByText('Обед')).toBeInTheDocument()
  })

  it('повторное нажатие в течение 5 секунд отменяет отметку', async () => {
    const user = setup()
    const fetchMock = mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'PATCH /api/tasks/1': () => jsonResponse({}),
    })
    render(<TodayTasks />)

    await user.click(await screen.findByLabelText('Отметить выполненной: Выпить таблетку'))
    act(() => vi.advanceTimersByTime(2000))
    await user.click(screen.getByLabelText('Отменить отметку: Выпить таблетку'))

    act(() => vi.advanceTimersByTime(10000))
    expect(screen.getByLabelText('Отметить выполненной: Выпить таблетку')).toBeInTheDocument()
    expect(patchCalls(fetchMock)).toEqual([
      ['/api/tasks/1', { done: true }],
      ['/api/tasks/1', { done: false }],
    ])
  })

  it('если сохранить не удалось, отметка снимается и задача не пропадает', async () => {
    const user = setup()
    mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'PATCH /api/tasks/1': () => jsonResponse({ error: 'задача не найдена' }, 404),
    })
    render(<TodayTasks />)

    await user.click(await screen.findByLabelText('Отметить выполненной: Выпить таблетку'))

    expect(await screen.findByText('Не удалось сохранить. Попробуйте ещё раз')).toBeInTheDocument()
    act(() => vi.advanceTimersByTime(6000))
    expect(screen.getByLabelText('Отметить выполненной: Выпить таблетку')).toBeInTheDocument()
  })

  it('можно отметить несколько задач подряд', async () => {
    const user = setup()
    const fetchMock = mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'PATCH /api/tasks/1': () => jsonResponse({}),
      'PATCH /api/tasks/3': () => jsonResponse({}),
    })
    render(<TodayTasks />)

    await user.click(await screen.findByLabelText('Отметить выполненной: Выпить таблетку'))
    await user.click(screen.getByLabelText('Отметить выполненной: Обед'))
    expect(patchCalls(fetchMock)).toHaveLength(2)

    act(() => vi.advanceTimersByTime(5100))
    expect(await screen.findByText('Все задачи на сегодня выполнены')).toBeInTheDocument()
  })

  it('двойной клик не шлёт второй запрос, пока первый не закончился', async () => {
    const user = setup()
    let finish
    const fetchMock = mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'PATCH /api/tasks/1': () =>
        new Promise((resolve) => {
          finish = () => resolve({ ok: true, status: 200, json: () => Promise.resolve({}) })
        }),
    })
    render(<TodayTasks />)

    const check = await screen.findByLabelText('Отметить выполненной: Выпить таблетку')
    await user.click(check)
    await user.click(check)
    await act(async () => finish())

    expect(patchCalls(fetchMock)).toHaveLength(1)
  })

  it('кнопка «+» открывает форму, новая задача встаёт по времени', async () => {
    const user = setup()
    const created = { id: 10, time: '10:00', title: 'Прогулка', kind: 'other', done: false }
    mockFetch({
      'GET /api/tasks/today': () => jsonResponse(openTasks),
      'POST /api/tasks': () => jsonResponse(created, 201),
    })
    render(<TodayTasks />)
    await screen.findByText('Выпить таблетку')

    await user.click(screen.getByLabelText('Добавить задачу'))
    const dialog = screen.getByRole('dialog', { name: 'Новая задача' })
    await user.type(screen.getByLabelText('Что нужно сделать'), 'Прогулка')
    await user.type(screen.getByLabelText('Время'), '10:00')
    await user.click(screen.getByRole('button', { name: 'Добавить' }))

    expect(dialog).not.toBeInTheDocument()
    const titles = screen.getAllByText(/Выпить таблетку|Прогулка|Обед/).map((el) => el.textContent)
    expect(titles).toEqual(['Выпить таблетку', 'Прогулка', 'Обед'])
  })

  it('задача со списком раскрывается, пункты отмечаются, «Где купить» ведёт к ценам', async () => {
    const user = userEvent.setup()
    const task = {
      id: 7,
      time: '10:00',
      title: 'Купить в аптеке',
      kind: 'medicine',
      done: false,
      note: 'Возьмите рецепт с собой',
      items: [
        { id: 1, title: 'Лизиноприл, 10 мг', medicineId: 'lizinopril', done: false },
        { id: 2, title: 'Тромбопол', medicineId: '', done: true },
      ],
    }
    let failNext = false
    const fetchMock = mockFetch({
      'GET /api/tasks/today': () => jsonResponse({ date: '2026-09-30', tasks: [task] }),
      'PATCH /api/tasks/7/items/1': () => (failNext ? jsonResponse({ error: 'нет' }, 500) : jsonResponse({ id: 1, done: true })),
    })
    const onOpenMedicine = vi.fn()
    render(<TodayTasks onOpenMedicine={onOpenMedicine} />)

    expect(await screen.findByText('10:00 · отмечено 1 из 2')).toBeInTheDocument()
    expect(screen.queryByText('Возьмите рецепт с собой')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /^Купить в аптеке/ }))
    expect(screen.getByRole('button', { name: /^Купить в аптеке/ })).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByText('Возьмите рецепт с собой')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Лизиноприл, 10 мг' }))
    expect(screen.getByRole('button', { name: 'Лизиноприл, 10 мг' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByText('10:00 · отмечено 2 из 2')).toBeInTheDocument()
    expect(JSON.parse(fetchMock.mock.calls[1][1].body)).toEqual({ done: true })

    failNext = true
    await user.click(screen.getByRole('button', { name: 'Лизиноприл, 10 мг' }))
    expect(await screen.findByText('Не удалось сохранить. Попробуйте ещё раз')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Лизиноприл, 10 мг' })).toHaveAttribute('aria-pressed', 'true')

    expect(screen.getAllByRole('button', { name: 'Где купить' })).toHaveLength(1)
    await user.click(screen.getByRole('button', { name: 'Где купить' }))
    expect(onOpenMedicine).toHaveBeenCalledWith({ id: 'lizinopril', name: '', form: '' })

    await user.click(screen.getByRole('button', { name: /^Купить в аптеке/ }))
    expect(screen.queryByText('Возьмите рецепт с собой')).not.toBeInTheDocument()
  })

  it('простая задача не раскрывается', async () => {
    mockFetch({ 'GET /api/tasks/today': () => jsonResponse(demoTasks) })
    render(<TodayTasks />)
    await screen.findByText('Выпить таблетку')
    expect(screen.queryByText('Подробнее')).not.toBeInTheDocument()
  })
})
