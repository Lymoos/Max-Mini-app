import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import AddTaskForm from './AddTaskForm'

describe('AddTaskForm', () => {
  it('кнопка «Добавить» неактивна, пока не заполнены текст и время', async () => {
    const user = userEvent.setup()
    render(<AddTaskForm onAdded={() => {}} onClose={() => {}} />)
    const addBtn = screen.getByRole('button', { name: 'Добавить' })

    expect(addBtn).toBeDisabled()
    await user.type(screen.getByLabelText('Что нужно сделать'), '   ')
    await user.type(screen.getByLabelText('Время'), '09:00')
    expect(addBtn).toBeDisabled()

    await user.type(screen.getByLabelText('Что нужно сделать'), 'Зарядка')
    expect(addBtn).toBeEnabled()
  })

  it('отправляет задачу с выбранным видом и закрывается', async () => {
    const user = userEvent.setup()
    const created = { id: 7, title: 'Аспирин', time: '21:00', kind: 'medicine', done: false }
    const fetchMock = mockFetch({ 'POST /api/tasks': () => jsonResponse(created, 201) })
    const onAdded = vi.fn()
    const onClose = vi.fn()
    render(<AddTaskForm onAdded={onAdded} onClose={onClose} />)

    expect(screen.getByRole('button', { name: 'Другое' })).toHaveAttribute('aria-pressed', 'true')
    await user.type(screen.getByLabelText('Что нужно сделать'), 'Аспирин')
    await user.type(screen.getByLabelText('Время'), '21:00')
    await user.click(screen.getByRole('button', { name: 'Лекарство' }))
    expect(screen.getByRole('button', { name: 'Лекарство' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Другое' })).toHaveAttribute('aria-pressed', 'false')

    await user.click(screen.getByRole('button', { name: 'Добавить' }))

    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ title: 'Аспирин', time: '21:00', kind: 'medicine' })
    expect(onAdded).toHaveBeenCalledWith(created)
    expect(onClose).toHaveBeenCalled()
  })

  it('показывает ошибку сервера и не закрывается', async () => {
    const user = userEvent.setup()
    mockFetch({ 'POST /api/tasks': () => jsonResponse({ error: 'слишком длинное название' }, 400) })
    const onClose = vi.fn()
    render(<AddTaskForm onAdded={() => {}} onClose={onClose} />)

    await user.type(screen.getByLabelText('Что нужно сделать'), 'Дело')
    await user.type(screen.getByLabelText('Время'), '10:00')
    await user.click(screen.getByRole('button', { name: 'Добавить' }))

    expect(await screen.findByText('слишком длинное название')).toBeInTheDocument()
    expect(onClose).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Добавить' })).toBeEnabled()
  })

  it('«Отмена» и клик по затемнению закрывают, клик внутри формы нет', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    const { container } = render(<AddTaskForm onAdded={() => {}} onClose={onClose} />)

    await user.click(screen.getByLabelText('Что нужно сделать'))
    expect(onClose).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: 'Отмена' }))
    expect(onClose).toHaveBeenCalledTimes(1)

    await user.click(container.querySelector('.overlay'))
    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('ограничивает длину названия', () => {
    render(<AddTaskForm onAdded={() => {}} onClose={() => {}} />)
    expect(screen.getByLabelText('Что нужно сделать')).toHaveAttribute('maxLength', '200')
  })
})
