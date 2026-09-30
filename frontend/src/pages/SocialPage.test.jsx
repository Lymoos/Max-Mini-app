import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import SocialPage from './SocialPage'

const location = { address: 'Мясницкая улица, 20, Москва', lat: 55.76, lon: 37.63, isDefault: false }
const places = [
  { id: 1, kind: 'social', name: 'ЦСО «Мещанский»', address: 'Сретенка, 5', phone: '+7 (495) 123-45-67', lat: 55.76, lon: 37.63, rating: 4.6, reviews: 21, distanceKm: 0.4 },
  { id: 2, kind: 'social', name: 'Совет ветеранов', address: '', phone: '', lat: 55.76, lon: 37.63, rating: 4.1, reviews: 1, distanceKm: 1.2 },
]

function renderSocial(routes) {
  mockFetch(routes)
  const onBack = vi.fn()
  const onOpenProfile = vi.fn()
  render(<SocialPage onBack={onBack} onOpenProfile={onOpenProfile} />)
  return { onBack, onOpenProfile }
}

describe('SocialPage', () => {
  it('показывает организации с телефоном, рейтингом и расстоянием', async () => {
    renderSocial({ 'GET /api/social/nearby': () => jsonResponse({ location, places }) })

    expect(await screen.findByText('ЦСО «Мещанский»')).toBeInTheDocument()
    expect(screen.getByText('+7 (495) 123-45-67')).toBeInTheDocument()
    expect(screen.getByText('Сретенка, 5 · 400 м')).toBeInTheDocument()
    expect(screen.getByText('21 отзыв', { exact: false })).toBeInTheDocument()
    expect(screen.getByText('Телефон не указан')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Совет ветеранов, телефон не указан' })).toBeDisabled()
  })

  it('нажатие открывает окно звонка с набранным номером', async () => {
    const user = userEvent.setup()
    renderSocial({ 'GET /api/social/nearby': () => jsonResponse({ location, places }) })

    await user.click(await screen.findByRole('button', { name: 'Позвонить: ЦСО «Мещанский»' }))
    const dialog = screen.getByRole('dialog', { name: 'Позвонить' })
    expect(within(dialog).getByRole('link', { name: /Позвонить/ })).toHaveAttribute('href', 'tel:+74951234567')
    expect(within(dialog).queryByText(/MAX/)).not.toBeInTheDocument()

    await user.click(within(dialog).getByRole('button', { name: 'Отмена' }))
    expect(screen.queryByRole('dialog', { name: 'Позвонить' })).not.toBeInTheDocument()
  })

  it('пусто, ошибка с повтором, назад и смена адреса', async () => {
    const user = userEvent.setup()
    const { onBack, onOpenProfile } = renderSocial({})
    expect(await screen.findByText('Не удалось загрузить организации')).toBeInTheDocument()

    mockFetch({ 'GET /api/social/nearby': () => jsonResponse({ location, places: [] }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Рядом организаций соцпомощи не нашли')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Изменить' }))
    expect(onOpenProfile).toHaveBeenCalled()
    await user.click(screen.getByText('Назад'))
    expect(onBack).toHaveBeenCalled()
  })
})
