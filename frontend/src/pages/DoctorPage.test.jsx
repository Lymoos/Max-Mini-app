import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import DoctorPage from './DoctorPage'

const doctor = { id: 12, clinicId: 2, name: 'Сафин Рустам Ренатович', specialty: 'Хирург', experience: 21, category: 'Высшая категория', rating: 4.9, reviews: 2 }
const clinic = { id: 2, kind: 'clinic', name: 'ГП № 2', address: 'Покровка, 1', lat: 55.76, lon: 37.64, rating: 4, reviews: 5, distanceKm: 0.8 }
const moscow = { region: 'moscow', title: 'Записаться в ЕМИАС', url: 'https://emias.info', phone: '122' }

function renderDoctor(routes) {
  const fetchMock = mockFetch({ 'GET /api/doctors/12': () => jsonResponse({ doctor, clinic, booking: moscow }), ...routes })
  render(<DoctorPage doctorId={12} onBack={() => {}} />)
  return fetchMock
}

describe('DoctorPage', () => {
  it('показывает врача и куда записываться', async () => {
    renderDoctor({})

    expect(await screen.findByRole('heading', { name: doctor.name })).toBeInTheDocument()
    expect(screen.getByText('Стаж 21 год · Высшая категория')).toBeInTheDocument()
    expect(screen.getByText('ГП № 2')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Записаться в ЕМИАС' })).toHaveAttribute('href', 'https://emias.info')
    expect(screen.getByRole('link', { name: 'Записаться в ЕМИАС' })).toHaveAttribute('target', '_blank')
    expect(screen.getByRole('link', { name: /Позвонить 122/ })).toHaveAttribute('href', 'tel:122')
    expect(screen.getByText('Данные врача — пример для демонстрации')).toBeInTheDocument()
  })

  it('«Я записался» создаёт напоминание на день приёма', async () => {
    const user = userEvent.setup()
    const fetchMock = renderDoctor({ 'POST /api/tasks': () => jsonResponse({ id: 1 }, 201) })

    await user.click(await screen.findByRole('button', { name: 'Я записался — напомнить' }))
    const dialog = screen.getByRole('dialog', { name: 'Напомнить о приёме' })
    const submit = within(dialog).getByRole('button', { name: 'Напомнить' })
    expect(submit).toBeDisabled()

    await user.type(within(dialog).getByLabelText('Дата'), '2026-10-02')
    expect(submit).toBeDisabled()
    await user.type(within(dialog).getByLabelText('Время'), '10:30')
    await user.click(submit)

    const post = fetchMock.mock.calls.find((c) => c[0] === '/api/tasks')
    expect(JSON.parse(post[1].body)).toEqual({ title: 'Приём: хирург, ГП № 2', time: '10:30', kind: 'doctor', date: '2026-10-02' })
    expect(await screen.findByText('Напомним о приёме 2 октября в 10:30')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('ошибка сохранения напоминания видна, окно остаётся', async () => {
    const user = userEvent.setup()
    renderDoctor({ 'POST /api/tasks': () => jsonResponse({ error: 'Дата должна быть не раньше сегодня и не дальше чем через год' }, 400) })

    await user.click(await screen.findByRole('button', { name: 'Я записался — напомнить' }))
    await user.type(screen.getByLabelText('Дата'), '2030-01-01')
    await user.type(screen.getByLabelText('Время'), '10:30')
    await user.click(screen.getByRole('button', { name: 'Напомнить' }))
    expect(await screen.findByText(/не дальше чем через год/)).toBeInTheDocument()
    expect(screen.getByRole('dialog', { name: 'Напомнить о приёме' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Отмена' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('ошибка загрузки и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({})
    render(<DoctorPage doctorId={12} onBack={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить врача')).toBeInTheDocument()

    mockFetch({ 'GET /api/doctors/12': () => jsonResponse({ doctor, clinic, booking: moscow }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByRole('heading', { name: doctor.name })).toBeInTheDocument()
  })
})
