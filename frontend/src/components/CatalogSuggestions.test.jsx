import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import { suggestMedicines } from '../api'
import CatalogSuggestions from './CatalogSuggestions'

const ibuprofen = { id: 'ibuprofen', name: 'Ибупрофен', form: 'таблетки', matched: 'Нурофен', corrected: true }
const paracetamol = { id: 'paracetamol', name: 'Парацетамол', form: 'таблетки', matched: 'Парацетамол', corrected: false }

function suggestUrl(q) {
  return 'GET /api/medicines/suggest?q=' + encodeURIComponent(q)
}

describe('CatalogSuggestions', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('ждёт 300 мс после ввода и показывает варианты', async () => {
    const fetchMock = mockFetch({ [suggestUrl('нурафен')]: () => jsonResponse([ibuprofen]) })
    render(<CatalogSuggestions query="нурафен" onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)

    act(() => vi.advanceTimersByTime(200))
    expect(fetchMock).not.toHaveBeenCalled()

    act(() => vi.advanceTimersByTime(150))
    expect(await screen.findByText('Ибупрофен')).toBeInTheDocument()
    expect(screen.getByText(/Нурофен/)).toBeInTheDocument()
    expect(screen.getByText('Возможно, вы имели в виду')).toBeInTheDocument()
    expect(screen.getByText('Найти в ближайших аптеках')).toBeInTheDocument()
  })

  it('без исправления нет подписи «вы имели в виду» и нет повтора названия', async () => {
    mockFetch({ [suggestUrl('пара')]: () => jsonResponse([paracetamol]) })
    render(<CatalogSuggestions query="пара" onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    act(() => vi.advanceTimersByTime(350))

    expect(await screen.findByText('Парацетамол')).toBeInTheDocument()
    expect(screen.queryByText('Возможно, вы имели в виду')).not.toBeInTheDocument()
    expect(screen.queryByText(/·/)).not.toBeInTheDocument()
  })

  it('короткий запрос не отправляется', () => {
    const fetchMock = mockFetch({})
    const { container } = render(<CatalogSuggestions query=" п " onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    act(() => vi.advanceTimersByTime(1000))
    expect(fetchMock).not.toHaveBeenCalled()
    expect(container).toBeEmptyDOMElement()
  })

  it('при быстром наборе уходит только последний запрос', async () => {
    const fetchMock = mockFetch({
      [suggestUrl('пар')]: () => jsonResponse([paracetamol]),
      [suggestUrl('пара')]: () => jsonResponse([paracetamol]),
    })
    const { rerender } = render(<CatalogSuggestions query="пар" onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    act(() => vi.advanceTimersByTime(100))
    rerender(<CatalogSuggestions query="пара" onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    act(() => vi.advanceTimersByTime(350))

    await screen.findByText('Парацетамол')
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock.mock.calls[0][0]).toBe('/api/medicines/suggest?q=' + encodeURIComponent('пара'))
  })

  it('нажатие на вариант отдаёт лекарство наверх', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    mockFetch({ [suggestUrl('нурафен')]: () => jsonResponse([ibuprofen]) })
    const onPick = vi.fn()
    render(<CatalogSuggestions query="нурафен" onPick={onPick} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    act(() => vi.advanceTimersByTime(350))

    await user.click(await screen.findByRole('button', { name: /Ибупрофен/ }))
    expect(onPick).toHaveBeenCalledWith(ibuprofen)
  })

  it('ошибка сервера и пустой ответ ничего не ломают', async () => {
    mockFetch({ [suggestUrl('абв')]: () => jsonResponse({ error: 'x' }, 500) })
    const { container } = render(<CatalogSuggestions query="абв" onPick={() => {}} fetcher={suggestMedicines} icon="pill" subtitle="Найти в ближайших аптеках" />)
    await act(async () => vi.advanceTimersByTime(350))
    expect(container).toBeEmptyDOMElement()
  })
})

describe('CatalogSuggestions: свои иконки', () => {
  it('иконка и подпись могут приходить у каждого варианта свои', async () => {
    const fetcher = vi.fn(() =>
      Promise.resolve([{ id: 'milk', name: 'Молоко', matched: 'Молоко', corrected: false, icon: 'cart', subtitle: 'Найти в магазинах рядом' }]),
    )
    render(<CatalogSuggestions query="мол" fetcher={fetcher} icon="pill" subtitle="по умолчанию" onPick={() => {}} />)

    expect(await screen.findByText('Найти в магазинах рядом', {}, { timeout: 2000 })).toBeInTheDocument()
    expect(screen.queryByText('по умолчанию')).not.toBeInTheDocument()
    expect(fetcher).toHaveBeenCalledWith('мол')
    expect(document.querySelector('.suggestion-icon-cart')).not.toBeNull()
  })
})
