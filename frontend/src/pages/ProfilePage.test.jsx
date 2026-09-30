import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import ProfilePage from './ProfilePage'

vi.mock('../components/MapPicker', () => ({
  default: ({ onPick }) => (
    <button type="button" onClick={() => onPick(55.757, 37.613)}>
      тестовая карта
    </button>
  ),
}))

const emptyProfile = {
  name: '',
  birthDate: '',
  address: '',
  lat: 0,
  lon: 0,
  hasLocation: false,
  contactName: '',
  contactPhone: '',
  health: '',
}

const place = { address: 'Тверская улица, 7, Москва', lat: 55.757, lon: 37.613 }

function renderProfile(routes) {
  const fetchMock = mockFetch(routes)
  const onClose = vi.fn()
  const onSaved = vi.fn()
  render(<ProfilePage onClose={onClose} onSaved={onSaved} />)
  return { fetchMock, onClose, onSaved }
}

async function homeBlock() {
  const title = await screen.findByRole('heading', { name: 'Адрес' })
  return within(title.closest('.profile-section'))
}

function bodies(fetchMock, path) {
  return fetchMock.mock.calls.filter((c) => c[0] === path && c[1] && c[1].method === 'PUT').map((c) => JSON.parse(c[1].body))
}

describe('ProfilePage', () => {
  it('загружает профиль, показывает дату рождения с возрастом', async () => {
    renderProfile({
      'GET /api/profile': () =>
        jsonResponse({ ...emptyProfile, name: 'Анна', birthDate: '1956-03-12', address: place.address, lat: 55.7, lon: 37.6, hasLocation: true }),
    })

    expect(await screen.findByLabelText('Как к вам обращаться')).toHaveValue('Анна')
    expect(screen.getByText('12 марта 1956')).toBeInTheDocument()
    expect(screen.getByText(/лет$/)).toBeInTheDocument()
    expect(screen.getByText(place.address)).toBeInTheDocument()
    expect(screen.queryByRole('spinbutton')).not.toBeInTheDocument()
  })

  it('кнопка «?» объясняет, зачем дата рождения', async () => {
    const user = userEvent.setup()
    renderProfile({ 'GET /api/profile': () => jsonResponse(emptyProfile) })

    const why = await screen.findByLabelText('Зачем указывать дату рождения')
    expect(screen.queryByText(/скидкой в ваш день рождения/)).not.toBeInTheDocument()
    await user.click(why)
    expect(screen.getByText(/скидкой в ваш день рождения/)).toBeInTheDocument()
    expect(screen.getByText(/менять паспорт/)).toBeInTheDocument()
    await user.click(why)
    expect(screen.queryByText(/скидкой в ваш день рождения/)).not.toBeInTheDocument()
  })

  it('дата выбирается в календаре и сохраняется', async () => {
    const user = userEvent.setup()
    const { fetchMock, onSaved, onClose } = renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'PUT /api/profile': () => jsonResponse(emptyProfile),
    })

    await user.click(await screen.findByRole('button', { name: 'Дата рождения' }))
    await user.click(screen.getByRole('button', { name: '12' }))
    await user.click(screen.getByRole('button', { name: 'Март' }))
    await user.click(screen.getByRole('button', { name: '1956' }))

    expect(screen.queryByRole('dialog', { name: 'Дата рождения' })).not.toBeInTheDocument()
    expect(screen.getByText('12 марта 1956')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Сохранить' }))
    expect(bodies(fetchMock, '/api/profile')[0].birthDate).toBe('1956-03-12')
    expect(onSaved).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })

  it('крестик сохраняет несохранённые изменения', async () => {
    const user = userEvent.setup()
    const { fetchMock, onClose } = renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'PUT /api/profile': () => jsonResponse(emptyProfile),
    })

    await user.type(await screen.findByLabelText('Как к вам обращаться'), 'Анна')
    await user.click(screen.getByLabelText('Закрыть'))

    expect(bodies(fetchMock, '/api/profile')).toEqual([{ ...emptyProfile, name: 'Анна' }])
    expect(onClose).toHaveBeenCalled()
  })

  it('крестик без изменений просто закрывает', async () => {
    const user = userEvent.setup()
    const { fetchMock, onClose } = renderProfile({ 'GET /api/profile': () => jsonResponse(emptyProfile) })
    await screen.findByLabelText('Как к вам обращаться')
    await user.click(screen.getByLabelText('Закрыть'))
    expect(onClose).toHaveBeenCalled()
    expect(bodies(fetchMock, '/api/profile')).toHaveLength(0)
  })

  it('если при закрытии сохранить не вышло, окно остаётся с ошибкой', async () => {
    const user = userEvent.setup()
    const { onClose } = renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'PUT /api/profile': () => jsonResponse({ error: 'Проверьте номер телефона' }, 400),
    })

    await user.type(await screen.findByLabelText('Телефон'), '123')
    await user.click(screen.getByLabelText('Закрыть'))

    expect(await screen.findByText('Проверьте номер телефона')).toBeInTheDocument()
    expect(onClose).not.toHaveBeenCalled()
  })

  it('выбранный адрес сохраняется сразу, без кнопки «Сохранить»', async () => {
    const user = userEvent.setup()
    const { fetchMock, onSaved, onClose } = renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      ['GET /api/geo/search?q=' + encodeURIComponent('Тверская 7')]: () => jsonResponse([place]),
      'PUT /api/profile/address': () => jsonResponse(place),
    })

    await user.type(await screen.findByLabelText('Поиск адреса'), 'Тверская 7{Enter}')
    await user.click(await screen.findByRole('button', { name: place.address }))

    expect(await screen.findByText('Адрес сохранён')).toBeInTheDocument()
    expect(bodies(fetchMock, '/api/profile/address')).toEqual([place])
    expect(onSaved).toHaveBeenCalled()

    await user.click(screen.getByLabelText('Закрыть'))
    expect(bodies(fetchMock, '/api/profile')).toHaveLength(0)
    expect(onClose).toHaveBeenCalled()
  })

  it('выбор на карте и удаление адреса тоже сохраняются сразу', async () => {
    const user = userEvent.setup()
    const { fetchMock } = renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'GET /api/geo/reverse?lat=55.757&lon=37.613': () => jsonResponse(place),
      'PUT /api/profile/address': () => jsonResponse(place),
    })

    await user.click((await homeBlock()).getByRole('button', { name: /Выбрать на карте/ }))
    await user.click(screen.getByText('тестовая карта'))
    expect(await screen.findByText(place.address)).toBeInTheDocument()

    await user.click(screen.getByLabelText('Убрать адрес'))
    expect(await screen.findByText('Адрес удалён')).toBeInTheDocument()
    expect(bodies(fetchMock, '/api/profile/address')).toEqual([place, { address: '' }])
  })

  it('ошибка сохранения адреса видна', async () => {
    const user = userEvent.setup()
    renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'GET /api/geo/reverse?lat=55.757&lon=37.613': () => jsonResponse(place),
      'PUT /api/profile/address': () => jsonResponse({ error: 'Что-то пошло не так' }, 500),
    })

    await user.click((await homeBlock()).getByRole('button', { name: /Выбрать на карте/ }))
    await user.click(screen.getByText('тестовая карта'))
    expect(await screen.findByText('Что-то пошло не так')).toBeInTheDocument()
    expect(screen.queryByText('Адрес сохранён')).not.toBeInTheDocument()
  })

  it('короткий и ненайденный адрес, сервис недоступен', async () => {
    const user = userEvent.setup()
    renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      ['GET /api/geo/search?q=' + encodeURIComponent('абвгд')]: () => jsonResponse([]),
      ['GET /api/geo/search?q=' + encodeURIComponent('абвгдe')]: () =>
        jsonResponse({ error: 'Сервис адресов сейчас недоступен' }, 502),
    })

    await user.type(await screen.findByLabelText('Поиск адреса'), 'аб')
    await user.click((await homeBlock()).getByRole('button', { name: 'Найти' }))
    expect(screen.getByText('Напишите адрес подробнее')).toBeInTheDocument()

    await user.type(screen.getByLabelText('Поиск адреса'), 'вгд')
    await user.click((await homeBlock()).getByRole('button', { name: 'Найти' }))
    expect(await screen.findByText('Адрес не найден. Попробуйте написать по-другому')).toBeInTheDocument()

    await user.type(screen.getByLabelText('Поиск адреса'), 'e')
    await user.click((await homeBlock()).getByRole('button', { name: 'Найти' }))
    expect(await screen.findByText('Сервис адресов сейчас недоступен')).toBeInTheDocument()
  })

  it('«Где я сейчас»: геолокация есть и нет', async () => {
    const user = userEvent.setup()
    renderProfile({
      'GET /api/profile': () => jsonResponse(emptyProfile),
      'GET /api/geo/reverse?lat=55.757&lon=37.613': () => jsonResponse(place),
      'PUT /api/profile/address': () => jsonResponse(place),
    })

    const saved = navigator.geolocation
    Object.defineProperty(navigator, 'geolocation', {
      configurable: true,
      value: { getCurrentPosition: (ok) => ok({ coords: { latitude: 55.757, longitude: 37.613 } }) },
    })
    await user.click((await homeBlock()).getByRole('button', { name: /Где я сейчас/ }))
    expect(await screen.findByText(place.address)).toBeInTheDocument()

    Object.defineProperty(navigator, 'geolocation', {
      configurable: true,
      value: { getCurrentPosition: (ok, fail) => fail({ code: 1 }) },
    })
    await user.click((await homeBlock()).getByRole('button', { name: /Где я сейчас/ }))
    expect(
      screen.getByText('Не удалось определить местоположение. Разрешите доступ или выберите на карте'),
    ).toBeInTheDocument()

    Object.defineProperty(navigator, 'geolocation', { configurable: true, value: saved })
  })

  it('ошибка загрузки и повтор', async () => {
    const user = userEvent.setup()
    renderProfile({})
    expect(await screen.findByText('Не удалось загрузить профиль')).toBeInTheDocument()

    mockFetch({ 'GET /api/profile': () => jsonResponse(emptyProfile) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByLabelText('Как к вам обращаться')).toBeInTheDocument()
  })

  it('прописка: «совпадает с адресом проживания» и поиск сохраняются сразу', async () => {
    const user = userEvent.setup()
    const withHome = { ...emptyProfile, address: place.address, lat: place.lat, lon: place.lon, hasLocation: true, regAddress: '', regLat: 0, regLon: 0, hasRegistration: false, clinicId: 0 }
    const other = { address: 'Казань, Баумана, 1', lat: 55.79, lon: 49.12 }
    const { fetchMock, onSaved } = renderProfile({
      'GET /api/profile': () => jsonResponse(withHome),
      'PUT /api/profile/registration': () => jsonResponse({ hasRegistration: true }),
      ['GET /api/geo/search?q=' + encodeURIComponent('Баумана 1')]: () => jsonResponse([other]),
    })

    await user.click(await screen.findByRole('button', { name: 'Совпадает с адресом проживания' }))
    expect(await screen.findByText('Прописка сохранена')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Совпадает с адресом проживания' })).not.toBeInTheDocument()
    expect(onSaved).toHaveBeenCalled()

    await user.type(screen.getByLabelText('Поиск адреса прописки'), 'Баумана 1{Enter}')
    await user.click(await screen.findByRole('button', { name: other.address }))
    await user.click(screen.getByLabelText('Убрать прописку'))

    expect(bodies(fetchMock, '/api/profile/registration')).toEqual([place, other, { address: '' }])
    expect(await screen.findByText('Прописка удалена')).toBeInTheDocument()
  })
})
