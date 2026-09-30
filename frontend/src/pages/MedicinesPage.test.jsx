import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import MedicinesPage from './MedicinesPage'

const defaultLocation = { address: 'Москва, центр', lat: 55.75, lon: 37.62, isDefault: true }
const homeLocation = { address: 'Мясницкая улица, 20, Москва', lat: 55.76, lon: 37.63, isDefault: false }

const pharmacies = [
  { id: 'p1', name: 'Аптека Здоровье', address: 'ул. Тверская, 8', lat: 55.76, lon: 37.6, rating: 4.8, reviews: 312, distanceKm: 1.08, score: 0.74 },
  { id: 'p2', name: 'Аптека №1', address: 'ул. Никольская, 10', lat: 55.75, lon: 37.62, rating: 4.2, reviews: 1, distanceKm: 0.53, score: 0.69 },
]

const paracetamol = { id: 'paracetamol', name: 'Парацетамол', form: 'таблетки 500 мг, 20 шт' }

const offers = [
  { ...pharmacies[0], price: 47, cheapest: false },
  { ...pharmacies[1], price: 45, cheapest: true },
]

function renderPage(props) {
  const handlers = {
    onPickMedicine: vi.fn(),
    onClearMedicine: vi.fn(),
    onOpenProfile: vi.fn(),
  }
  render(<MedicinesPage medicine={null} {...handlers} {...props} />)
  return handlers
}

describe('MedicinesPage', () => {
  it('без лекарства показывает лучшие аптеки рядом', async () => {
    mockFetch({ 'GET /api/pharmacies/nearby': () => jsonResponse({ location: defaultLocation, pharmacies }) })
    const handlers = renderPage()

    expect(await screen.findByText('Аптека Здоровье')).toBeInTheDocument()
    expect(screen.getByText('Лучшие аптеки рядом')).toBeInTheDocument()
    expect(screen.getByText(/ул\. Тверская, 8 · 1,1 км/)).toBeInTheDocument()
    expect(screen.getByText(/ул\. Никольская, 10 · 530 м/)).toBeInTheDocument()
    expect(screen.getByText(/1 отзыв$/)).toBeInTheDocument()
    expect(screen.queryByText(/₽/)).not.toBeInTheDocument()

    const names = screen.getAllByText(/^Аптека/).map((el) => el.textContent)
    expect(names).toEqual(['Аптека Здоровье', 'Аптека №1'])

    expect(screen.getByText('Адрес не указан, показываем центр Москвы')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Указать' }))
    expect(handlers.onOpenProfile).toHaveBeenCalled()
  })

  it('ссылка «Маршрут» ведёт в Яндекс Карты и открывается в новой вкладке', async () => {
    mockFetch({ 'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies }) })
    renderPage()

    const links = await screen.findAllByRole('link', { name: 'Маршрут' })
    expect(links[0]).toHaveAttribute('href', 'https://yandex.ru/maps/?rtext=~55.76,37.6&rtt=pd')
    expect(links[0]).toHaveAttribute('target', '_blank')
    expect(screen.getByText('Рядом с: Мясницкая улица, 20, Москва')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Изменить' })).toBeInTheDocument()
  })

  it('с лекарством показывает цены и помечает самую низкую', async () => {
    const user = userEvent.setup()
    mockFetch({
      'GET /api/medicines/paracetamol/offers': () =>
        jsonResponse({ medicine: paracetamol, location: homeLocation, offers }),
    })
    const handlers = renderPage({ medicine: paracetamol })

    expect(await screen.findByText('47 ₽')).toBeInTheDocument()
    expect(screen.getByText('таблетки 500 мг, 20 шт')).toBeInTheDocument()
    const cheapCard = screen.getByText('45 ₽').closest('li')
    expect(within(cheapCard).getByText('Дешевле всего')).toBeInTheDocument()
    expect(screen.getAllByText('Дешевле всего')).toHaveLength(1)
    expect(screen.queryByText('Лучшие аптеки рядом')).not.toBeInTheDocument()

    await user.click(screen.getByLabelText('Показать все аптеки'))
    expect(handlers.onClearMedicine).toHaveBeenCalled()
  })

  it('пустые списки', async () => {
    mockFetch({
      'GET /api/medicines/paracetamol/offers': () =>
        jsonResponse({ medicine: paracetamol, location: homeLocation, offers: [] }),
    })
    renderPage({ medicine: paracetamol })
    expect(await screen.findByText('Рядом нет аптек, где есть это лекарство')).toBeInTheDocument()
  })

  it('пустой список аптек', async () => {
    mockFetch({ 'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies: [] }) })
    renderPage()
    expect(await screen.findByText('Рядом аптек не нашли')).toBeInTheDocument()
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({})
    renderPage()
    expect(await screen.findByText('Не удалось загрузить аптеки')).toBeInTheDocument()

    mockFetch({ 'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies }) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Аптека Здоровье')).toBeInTheDocument()
  })

  it('Enter в поиске выбирает первое найденное лекарство', async () => {
    const user = userEvent.setup()
    mockFetch({
      'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies }),
      ['GET /api/medicines/suggest?q=' + encodeURIComponent('парацетомол')]: () =>
        jsonResponse([{ ...paracetamol, matched: 'Парацетамол', corrected: true }]),
    })
    const handlers = renderPage()

    await user.type(screen.getByLabelText('Название лекарства'), 'парацетомол{Enter}')
    expect(handlers.onPickMedicine).toHaveBeenCalledWith(expect.objectContaining({ id: 'paracetamol' }))
    expect(screen.getByLabelText('Название лекарства')).toHaveValue('')
  })

  it('если лекарство не найдено, пишет об этом', async () => {
    const user = userEvent.setup()
    mockFetch({
      'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies }),
      ['GET /api/medicines/suggest?q=' + encodeURIComponent('стол')]: () => jsonResponse([]),
    })
    const handlers = renderPage()

    await user.type(screen.getByLabelText('Название лекарства'), 'стол{Enter}')
    expect(await screen.findByText('Такое лекарство не нашли. Проверьте название')).toBeInTheDocument()
    expect(handlers.onPickMedicine).not.toHaveBeenCalled()

    await user.type(screen.getByLabelText('Название лекарства'), 'ы')
    expect(screen.queryByText('Такое лекарство не нашли. Проверьте название')).not.toBeInTheDocument()
  })

  it('пустой поиск не отправляется', async () => {
    const user = userEvent.setup()
    const fetchMock = mockFetch({
      'GET /api/pharmacies/nearby': () => jsonResponse({ location: homeLocation, pharmacies }),
    })
    renderPage()
    await screen.findByText('Аптека Здоровье')

    await user.type(screen.getByLabelText('Название лекарства'), '   {Enter}')
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })
})
