import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import GoodsPage from './GoodsPage'

const location = { address: 'Мясницкая улица, 20, Москва', lat: 55.76, lon: 37.63, isDefault: false }
const shops = [
  { id: 1, kind: 'shop', name: 'Магнолия', address: '', lat: 55.76, lon: 37.63, rating: 4.6, reviews: 231, distanceKm: 0.23, score: 0.8 },
  { id: 2, kind: 'shop', name: 'Дикси', address: 'Покровка, 3', lat: 55.75, lon: 37.64, rating: 4.1, reviews: 12, distanceKm: 1.4, score: 0.5 },
]
const milk = { id: 'milk', name: 'Молоко 2,5%', unit: '1 л' }

function renderGoods(props, routes) {
  const fetchMock = mockFetch(routes)
  const handlers = {
    onPickProduct: vi.fn(),
    onClearProduct: vi.fn(),
    onOpenShop: vi.fn(),
    onOpenProfile: vi.fn(),
    onBack: vi.fn(),
  }
  render(<GoodsPage product={null} {...handlers} {...props} />)
  return { fetchMock, ...handlers }
}

describe('GoodsPage', () => {
  it('показывает лучшие магазины и открывает магазин по нажатию', async () => {
    const user = userEvent.setup()
    const { onOpenShop, onBack } = renderGoods({}, { 'GET /api/shops/nearby': () => jsonResponse({ location, shops }) })

    expect(await screen.findByText('Магнолия')).toBeInTheDocument()
    expect(screen.getByText('Лучшие магазины рядом')).toBeInTheDocument()
    expect(screen.getByText('230 м')).toBeInTheDocument()
    expect(screen.getByText('Покровка, 3 · 1,4 км')).toBeInTheDocument()
    expect(screen.getByText('Рядом с: ' + location.address)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Дикси/ }))
    expect(onOpenShop).toHaveBeenCalledWith(shops[1])

    await user.click(screen.getByText('Назад'))
    expect(onBack).toHaveBeenCalled()
  })

  it('с товаром показывает цены, акции и самую низкую цену', async () => {
    const user = userEvent.setup()
    const offers = [
      { ...shops[0], price: 70, oldPrice: 100, promoUntil: '2026-10-03', cheapest: true },
      { ...shops[1], price: 95, cheapest: false },
    ]
    const { onClearProduct } = renderGoods(
      { product: milk },
      { 'GET /api/products/milk/offers': () => jsonResponse({ product: milk, location, offers }) },
    )

    const first = (await screen.findByText('70 ₽')).closest('li')
    expect(within(first).getByText('100 ₽')).toHaveClass('old-price')
    expect(within(first).getByText('Акция до 3 октября')).toBeInTheDocument()
    expect(within(first).getByText('Дешевле всего')).toBeInTheDocument()
    const second = screen.getByText('95 ₽').closest('li')
    expect(within(second).queryByText('Дешевле всего')).not.toBeInTheDocument()
    expect(screen.getByText('1 л')).toBeInTheDocument()

    await user.click(screen.getByLabelText('Показать все магазины'))
    expect(onClearProduct).toHaveBeenCalled()
  })

  it('поиск товара: Enter берёт первый найденный, пустой результат — сообщение', async () => {
    const user = userEvent.setup()
    const { onPickProduct } = renderGoods(
      {},
      {
        'GET /api/shops/nearby': () => jsonResponse({ location, shops }),
        ['GET /api/products/suggest?q=' + encodeURIComponent('малако')]: () =>
          jsonResponse([{ ...milk, matched: 'Молоко', corrected: true }]),
        ['GET /api/products/suggest?q=' + encodeURIComponent('стол')]: () => jsonResponse([]),
      },
    )

    await user.type(screen.getByLabelText('Название товара'), 'стол{Enter}')
    expect(await screen.findByText('Такой товар не нашли. Попробуйте написать по-другому')).toBeInTheDocument()

    await user.clear(screen.getByLabelText('Название товара'))
    await user.type(screen.getByLabelText('Название товара'), 'малако{Enter}')
    expect(onPickProduct).toHaveBeenCalledWith(expect.objectContaining({ id: 'milk' }))
  })

  it('нет магазинов с этим товаром', async () => {
    renderGoods({ product: milk }, {
      'GET /api/products/milk/offers': () => jsonResponse({ product: milk, location, offers: [] }),
    })
    expect(await screen.findByText('Рядом нет магазинов, где есть этот товар')).toBeInTheDocument()
  })

  it('ошибка загрузки и повтор', async () => {
    const user = userEvent.setup()
    renderGoods({}, {})
    expect(await screen.findByText('Не удалось загрузить магазины')).toBeInTheDocument()

    mockFetch({ 'GET /api/shops/nearby': () => jsonResponse({ location, shops: [] }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Рядом магазинов не нашли')).toBeInTheDocument()
  })
})
