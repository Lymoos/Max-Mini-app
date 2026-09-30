import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import ShopPage from './ShopPage'

const shop = {
  id: 7, kind: 'shop', name: 'Магнолия', address: 'Мясницкая, 1', lat: 55.76, lon: 37.63,
  rating: 4.6, reviews: 231, birthdayDiscount: 10, distanceKm: 0.23, score: 0.8,
}
const promos = [
  { productId: 'orange', name: 'Апельсины', unit: '1 кг', price: 91, oldPrice: 152, discount: 40, promoUntil: '2026-10-03' },
  { productId: 'butter', name: 'Масло сливочное 82%', unit: '180 г', price: 122, oldPrice: 193, discount: 36, promoUntil: '2026-10-06' },
]

describe('ShopPage', () => {
  it('показывает магазин, скидку в день рождения и акции', async () => {
    const user = userEvent.setup()
    const fetchMock = mockFetch({ 'GET /api/shops/7': () => jsonResponse({ shop, promos }) })
    const onBack = vi.fn()
    render(<ShopPage shopId={7} onBack={onBack} />)

    expect(await screen.findByRole('heading', { name: 'Магнолия' })).toBeInTheDocument()
    expect(screen.getByText('Мясницкая, 1 · 230 м')).toBeInTheDocument()
    expect(screen.getByText('Скидка 10% в день рождения')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Построить маршрут' })).toHaveAttribute('href', 'https://yandex.ru/maps/?rtext=~55.76,37.63&rtt=pd')

    const orange = screen.getByText('Апельсины').closest('li')
    expect(within(orange).getByText('−40%')).toBeInTheDocument()
    expect(within(orange).getByText('91 ₽')).toBeInTheDocument()
    expect(within(orange).getByText('152 ₽')).toHaveClass('old-price')
    expect(within(orange).getByText('1 кг · до 3 октября')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith('/api/shops/7', undefined)

    await user.click(screen.getByText('Назад'))
    expect(onBack).toHaveBeenCalled()
  })

  it('без акций и без скидки ко дню рождения', async () => {
    mockFetch({ 'GET /api/shops/7': () => jsonResponse({ shop: { ...shop, birthdayDiscount: 0 }, promos: [] }) })
    render(<ShopPage shopId={7} onBack={() => {}} />)
    expect(await screen.findByText('Сейчас акций нет')).toBeInTheDocument()
    expect(screen.queryByText(/в день рождения/)).not.toBeInTheDocument()
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/shops/7': () => jsonResponse({ error: 'Магазин не найден' }, 404) })
    render(<ShopPage shopId={7} onBack={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить магазин')).toBeInTheDocument()

    mockFetch({ 'GET /api/shops/7': () => jsonResponse({ shop, promos }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Апельсины')).toBeInTheDocument()
  })
})
