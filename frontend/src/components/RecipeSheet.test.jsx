import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import AskBox from './AskBox'
import { suggestVisit } from '../format'
import RecipeSheet from './RecipeSheet'

const photo = new File(['фото'], 'recipe.jpg', { type: 'image/jpeg' })
const found = {
  items: [
    { title: 'Лизиноприл', dose: '10 мг, по 1 таб. утром', medicineId: 'lizinopril' },
    { title: 'Тромбопол', dose: '', medicineId: '' },
    { title: 'Омепразол', dose: '20 мг', medicineId: 'omeprazol' },
  ],
}

describe('RecipeSheet', () => {
  it('ближайший час для похода в аптеку, поздно вечером — завтра утром', () => {
    expect(suggestVisit(new Date(2026, 8, 30, 14, 20))).toEqual({ date: '2026-09-30', time: '15:00', tomorrow: false })
    expect(suggestVisit(new Date(2026, 8, 30, 6, 5))).toEqual({ date: '2026-09-30', time: '08:00', tomorrow: false })
    expect(suggestVisit(new Date(2026, 8, 30, 21, 40))).toEqual({ date: '2026-10-01', time: '10:00', tomorrow: true })
    expect(suggestVisit(new Date(2026, 11, 31, 22, 0))).toEqual({ date: '2027-01-01', time: '10:00', tomorrow: true })
  })

  it('читает фото, можно снять лишнее и добавить список одной задачей', async () => {
    const user = userEvent.setup()
    let sentPhoto = null
    const created = { id: 5, title: 'Купить в аптеке', time: '18:30', kind: 'medicine', done: false, items: [] }
    const fetchMock = mockFetch({
      'POST /api/recipe': (options) => {
        sentPhoto = options.body.get('photo')
        return jsonResponse(found)
      },
      'POST /api/tasks': () => jsonResponse(created, 201),
    })
    const onAdded = vi.fn()
    render(<RecipeSheet file={photo} onAdded={onAdded} onClose={() => {}} onRetry={() => {}} />)

    expect(screen.getByText('Читаю рецепт…')).toBeInTheDocument()
    expect(await screen.findByText('Нашёл в рецепте')).toBeInTheDocument()
    expect(sentPhoto).toBe(photo)
    expect(screen.getByText('10 мг, по 1 таб. утром')).toBeInTheDocument()

    await user.click(screen.getByText('Омепразол'))
    expect(screen.getByRole('checkbox', { name: /Омепразол/ })).not.toBeChecked()

    const time = screen.getByLabelText(/Когда пойти в аптеку/)
    await user.clear(time)
    await user.type(time, '18:30')
    await user.click(screen.getByRole('button', { name: 'Добавить список в задачи' }))

    const body = JSON.parse(fetchMock.mock.calls[1][1].body)
    expect(body.title).toBe('Купить в аптеке')
    expect(body.kind).toBe('medicine')
    expect(body.time).toBe('18:30')
    expect(body.note).toContain('Возьмите рецепт')
    expect(body.items).toEqual([
      { title: 'Лизиноприл, 10 мг, по 1 таб. утром', medicineId: 'lizinopril' },
      { title: 'Тромбопол', medicineId: '' },
    ])
    expect(onAdded).toHaveBeenCalledWith(created, expect.any(Boolean))
  })

  it('если всё снято — добавить нельзя', async () => {
    const user = userEvent.setup()
    mockFetch({ 'POST /api/recipe': () => jsonResponse({ items: [found.items[0]] }) })
    render(<RecipeSheet file={photo} onAdded={() => {}} onClose={() => {}} onRetry={() => {}} />)

    await user.click(await screen.findByText('Лизиноприл'))
    expect(screen.getByRole('button', { name: 'Добавить список в задачи' })).toBeDisabled()
  })

  it('ничего не нашёл или ошибка — подсказка и «Другое фото»', async () => {
    const user = userEvent.setup()
    mockFetch({ 'POST /api/recipe': () => jsonResponse({ items: [] }) })
    const onRetry = vi.fn()
    const { unmount } = render(<RecipeSheet file={photo} onAdded={() => {}} onClose={() => {}} onRetry={onRetry} />)
    expect(await screen.findByText(/Не нашёл лекарств на фото/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Другое фото' }))
    expect(onRetry).toHaveBeenCalled()
    unmount()

    mockFetch({ 'POST /api/recipe': () => jsonResponse({ error: 'Не получилось прочитать рецепт. Попробуйте ещё раз' }, 502) })
    const onClose = vi.fn()
    render(<RecipeSheet file={photo} onAdded={() => {}} onClose={onClose} onRetry={() => {}} />)
    expect(await screen.findByText('Не получилось прочитать рецепт. Попробуйте ещё раз')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Закрыть' }))
    expect(onClose).toHaveBeenCalled()
  })
})

describe('кнопка рецепта на главной', () => {
  it('иконка сразу открывает выбор фото, после выбора — шторка, после добавления — сообщение', async () => {
    const user = userEvent.setup()
    mockFetch({
      'POST /api/recipe': () => jsonResponse({ items: [found.items[0]] }),
      'POST /api/tasks': () => jsonResponse({ id: 5, title: 'Купить в аптеке', time: '18:00', kind: 'medicine', done: false, items: [] }, 201),
    })
    const onTaskAdded = vi.fn()
    render(<AskBox onAnswer={() => {}} onTaskAdded={onTaskAdded} />)

    const input = screen.getByTestId('recipe-file')
    const clickSpy = vi.spyOn(input, 'click')
    await user.click(screen.getByRole('button', { name: 'Прочитать рецепт по фото' }))
    expect(clickSpy).toHaveBeenCalled()
    expect(input).toHaveAttribute('accept', 'image/*,application/pdf')

    await user.upload(input, photo)
    expect(screen.getByRole('dialog', { name: 'Рецепт по фото' })).toBeInTheDocument()
    await user.click(await screen.findByRole('button', { name: 'Добавить список в задачи' }))

    expect(await screen.findByText(/Список добавлен в задачи: «Купить в аптеке»/)).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(onTaskAdded).toHaveBeenCalled()
  })
})
