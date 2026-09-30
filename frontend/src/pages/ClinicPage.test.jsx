import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import ClinicPage from './ClinicPage'

const clinic = { id: 2, kind: 'clinic', name: 'Городская поликлиника № 2', address: 'Покровка, 1', phone: '+7 495 111-22-33', lat: 55.76, lon: 37.64, rating: 4, reviews: 5, distanceKm: 0.8 }
const doctors = [
  { id: 11, clinicId: 2, name: 'Иванова Анна Сергеевна', specialty: 'Терапевт', experience: 12, category: 'Первая категория', rating: 4.7, reviews: 31 },
  { id: 12, clinicId: 2, name: 'Сафин Рустам Ренатович', specialty: 'Хирург', experience: 21, category: 'Высшая категория', rating: 4.9, reviews: 2 },
  { id: 13, clinicId: 2, name: 'Петров Олег Иванович', specialty: 'Терапевт', experience: 3, category: '', rating: 4.2, reviews: 0 },
]

describe('ClinicPage', () => {
  it('показывает врачей со стажем и категорией, фильтрует по специальности', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/clinics/2': () => jsonResponse({ clinic, doctors, specialties: ['Терапевт', 'Хирург'] }) })
    const onOpenDoctor = vi.fn()
    render(<ClinicPage clinicId={2} onOpenDoctor={onOpenDoctor} onBack={() => {}} />)

    expect(await screen.findByRole('heading', { name: clinic.name })).toBeInTheDocument()
    expect(screen.getByText('Список врачей — пример для демонстрации')).toBeInTheDocument()
    expect(screen.getByText('Стаж 12 лет · Первая категория')).toBeInTheDocument()
    expect(screen.getByText('Стаж 3 года')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '+7 495 111-22-33' })).toHaveAttribute('href', 'tel:+74951112233')
    expect(screen.getAllByRole('button', { name: /Терапевт|Хирург/ }).filter((b) => b.classList.contains('doctor'))).toHaveLength(3)

    await user.click(screen.getByRole('button', { name: 'Хирург' }))
    expect(screen.getByRole('button', { name: 'Хирург' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.queryByText('Иванова Анна Сергеевна')).not.toBeInTheDocument()
    await user.click(screen.getByText('Сафин Рустам Ренатович'))
    expect(onOpenDoctor).toHaveBeenCalledWith(12)

    await user.click(screen.getByRole('button', { name: 'Все' }))
    expect(screen.getByText('Иванова Анна Сергеевна')).toBeInTheDocument()
  })

  it('без врачей и без телефона', async () => {
    mockFetch({ 'GET /api/clinics/2': () => jsonResponse({ clinic: { ...clinic, phone: '' }, doctors: [], specialties: [] }) })
    render(<ClinicPage clinicId={2} onOpenDoctor={() => {}} onBack={() => {}} />)
    expect(await screen.findByText('Врачей не нашли')).toBeInTheDocument()
    expect(screen.queryByRole('group', { name: 'Специальность' })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /\+7/ })).not.toBeInTheDocument()
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/clinics/2': () => jsonResponse({ error: 'Поликлиника не найдена' }, 404) })
    render(<ClinicPage clinicId={2} onOpenDoctor={() => {}} onBack={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить поликлинику')).toBeInTheDocument()

    mockFetch({ 'GET /api/clinics/2': () => jsonResponse({ clinic, doctors, specialties: ['Терапевт', 'Хирург'] }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Сафин Рустам Ренатович')).toBeInTheDocument()
  })
})
