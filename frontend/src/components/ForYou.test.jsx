import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import ForYou from './ForYou'

const items = [
  {
    type: 'birthday',
    title: 'До дня рождения 5 дней',
    text: 'В этих местах рядом скидка именинникам.',
    places: [
      { id: 1, kind: 'shop', name: 'ВкусВилл', birthdayDiscount: 7, distanceKm: 1.7 },
      { id: 2, kind: 'pharmacy', name: '36,6', birthdayDiscount: 10, distanceKm: 0.1 },
    ],
  },
  { type: 'passport', title: 'Скоро менять паспорт', text: '5 октября 2026 вам исполнится 45.' },
]

describe('ForYou', () => {
  it('показывает день рождения со скидками и напоминание о паспорте', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/for-you': () => jsonResponse(items) })
    const onOpenShop = vi.fn()
    render(<ForYou onOpenShop={onOpenShop} />)

    expect(await screen.findByText('До дня рождения 5 дней')).toBeInTheDocument()
    expect(screen.getByText('Скоро менять паспорт')).toBeInTheDocument()
    expect(screen.getByText('−7% · 1,7 км')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /ВкусВилл/ }))
    expect(onOpenShop).toHaveBeenCalledWith(items[0].places[0])
    expect(screen.queryByRole('button', { name: /36,6/ })).not.toBeInTheDocument()
    expect(screen.getByText('36,6')).toBeInTheDocument()
  })

  it('ничего не рисует, если подсказок нет или сервер упал', async () => {
    mockFetch({ 'GET /api/for-you': () => jsonResponse([]) })
    const { container, unmount } = render(<ForYou onOpenShop={() => {}} />)
    await new Promise((r) => setTimeout(r, 50))
    expect(container).toBeEmptyDOMElement()
    unmount()

    mockFetch({})
    const second = render(<ForYou onOpenShop={() => {}} />)
    await new Promise((r) => setTimeout(r, 50))
    expect(second.container).toBeEmptyDOMElement()
  })
})
