import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { demoFeatures, demoTasks, jsonResponse, mockFetch } from './testUtils'

const location = { address: 'Москва, центр', lat: 55.75, lon: 37.62, isDefault: true }
const pharmacy = { id: 'p1', name: 'Аптека Здоровье', address: 'ул. Тверская, 8', lat: 55.76, lon: 37.6, rating: 4.8, reviews: 312, distanceKm: 1.08, score: 0.74 }
const ibuprofen = { id: 'ibuprofen', name: 'Ибупрофен', form: 'таблетки 200 мг, 20 шт' }

function mockAll(askAnswer) {
  return mockFetch({
    'GET /api/tasks/today': () => jsonResponse(demoTasks),
    'GET /api/features': () => jsonResponse(demoFeatures),
    'POST /api/ask': () => jsonResponse(askAnswer),
    'GET /api/pharmacies/nearby': () => jsonResponse({ location: location, pharmacies: [pharmacy] }),
    'GET /api/medicines/ibuprofen/offers': () =>
      jsonResponse({ medicine: ibuprofen, location: location, offers: [{ ...pharmacy, price: 70, cheapest: true }] }),
    'GET /api/profile': () =>
      jsonResponse({ name: '', age: 0, address: '', lat: 0, lon: 0, hasLocation: false, contactName: '', contactPhone: '', health: '' }),
    'PUT /api/profile': () => jsonResponse({}),
    'GET /api/for-you': () => jsonResponse([]),
    'GET /api/shops/nearby': () => jsonResponse({ location: location, shops: [shop] }),
    'GET /api/shops/5': () => jsonResponse({ shop: shop, promos: [] }),
    'GET /api/products/milk/offers': () =>
      jsonResponse({ product: milk, location: location, offers: [{ ...shop, price: 80, oldPrice: 100, cheapest: true }] }),
  })
}

const shop = { id: 5, kind: 'shop', name: 'Магнолия', address: '', lat: 55.76, lon: 37.63, rating: 4.6, reviews: 231, distanceKm: 0.2, score: 0.8 }
const milk = { id: 'milk', name: 'Молоко 2,5%', unit: '1 л' }

describe('App', () => {
  beforeEach(() => {
    window.scrollTo = vi.fn()
  })

  it('переключает вкладки нижнего меню', async () => {
    const user = userEvent.setup()
    mockAll({})
    render(<App />)
    const nav = screen.getByRole('navigation')

    expect(within(nav).getByText('Дом').closest('button')).toHaveAttribute('aria-current', 'page')

    await user.click(within(nav).getByText('Документы'))
    expect(screen.getByRole('heading', { name: 'Документы и льготы' })).toBeInTheDocument()
    expect(screen.queryByText('Задачи на сегодня')).not.toBeInTheDocument()
    expect(within(nav).getByText('Документы').closest('button')).toHaveAttribute('aria-current', 'page')

    await user.click(within(nav).getByText('Дом'))
    expect(await screen.findByText('Выпить таблетку')).toBeInTheDocument()
  })

  it('поиск открывает вкладку', async () => {
    const user = userEvent.setup()
    mockAll({ type: 'tab', target: 'documents', message: 'Открываю' })
    render(<App />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'паспорт{Enter}')
    expect(await screen.findByRole('heading', { name: 'Документы и льготы' })).toBeInTheDocument()
  })

  it('поиск лекарства открывает цены в аптеках, крестик — все аптеки', async () => {
    const user = userEvent.setup()
    mockAll({ type: 'medicine', target: 'ibuprofen', message: 'Ищем', medicine: ibuprofen })
    render(<App />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'где купить нурофен{Enter}')
    expect(await screen.findByText('70 ₽')).toBeInTheDocument()
    expect(screen.getByText('Ибупрофен')).toBeInTheDocument()
    expect(within(screen.getByRole('navigation')).queryByText('Лекарства')).not.toBeInTheDocument()

    await user.click(screen.getByLabelText('Показать все аптеки'))
    expect(await screen.findByText('Лучшие аптеки рядом')).toBeInTheDocument()
  })

  it('плитка «Аптеки рядом» открывает раздел лекарств', async () => {
    const user = userEvent.setup()
    mockAll({})
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Аптеки рядом' }))
    expect(await screen.findByText('Лучшие аптеки рядом')).toBeInTheDocument()
    expect(screen.getByText('Аптека Здоровье')).toBeInTheDocument()
  })

  it('профиль открывается поверх экрана, после сохранения аптеки перезагружаются', async () => {
    const user = userEvent.setup()
    const fetchMock = mockAll({})
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Аптеки рядом' }))
    await screen.findByText('Аптека Здоровье')
    await user.click(screen.getByRole('button', { name: 'Указать' }))

    expect(screen.getByRole('dialog', { name: 'Профиль' })).toBeInTheDocument()
    await user.click(await screen.findByRole('button', { name: 'Сохранить' }))

    expect(screen.queryByRole('dialog', { name: 'Профиль' })).not.toBeInTheDocument()
    await screen.findByText('Аптека Здоровье')
    const nearbyCalls = fetchMock.mock.calls.filter((c) => c[0] === '/api/pharmacies/nearby')
    expect(nearbyCalls.length).toBeGreaterThanOrEqual(2)
  })

  it('кнопка профиля на главной', async () => {
    const user = userEvent.setup()
    mockAll({})
    render(<App />)

    await user.click(screen.getByLabelText('Профиль'))
    expect(screen.getByRole('dialog', { name: 'Профиль' })).toBeInTheDocument()
    await user.click(screen.getByLabelText('Закрыть'))
    expect(screen.queryByRole('dialog', { name: 'Профиль' })).not.toBeInTheDocument()
  })

  it('поиск открывает возможность, «Назад» возвращает на главную', async () => {
    const user = userEvent.setup()
    mockAll({ type: 'feature', target: 'goods', message: 'Открываю', feature: demoFeatures[1] })
    render(<App />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'магазин{Enter}')
    expect(await screen.findByRole('heading', { name: 'Товары рядом' })).toBeInTheDocument()

    await user.click(screen.getByText('Назад'))
    expect(await screen.findByText('Задачи на сегодня')).toBeInTheDocument()
  })

  it('плитка открывает возможность, нижнее меню закрывает её', async () => {
    const user = userEvent.setup()
    mockAll({})
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Товары рядом' }))
    expect(screen.getByRole('heading', { name: 'Товары рядом' })).toBeInTheDocument()

    await user.click(within(screen.getByRole('navigation')).getByText('Помощь'))
    expect(screen.getByRole('heading', { name: 'Помощь' })).toBeInTheDocument()
  })

  it('товары рядом → магазин → назад → главная', async () => {
    const user = userEvent.setup()
    mockAll({})
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Товары рядом' }))
    await user.click(await screen.findByRole('button', { name: /Магнолия/ }))
    expect(await screen.findByText('Сейчас акций нет')).toBeInTheDocument()

    await user.click(screen.getByText('Назад'))
    expect(await screen.findByText('Лучшие магазины рядом')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))
    expect(await screen.findByText('Задачи на сегодня')).toBeInTheDocument()
  })

  it('поиск товара открывает цены в магазинах', async () => {
    const user = userEvent.setup()
    mockAll({ type: 'product', target: 'milk', message: 'Ищем', product: milk })
    render(<App />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'нужно молоко{Enter}')
    expect(await screen.findByText('80 ₽')).toBeInTheDocument()
    expect(screen.getByText('Молоко 2,5%')).toBeInTheDocument()
  })

  it('после смены адреса в профиле аптеки показываются от нового адреса', async () => {
    const user = userEvent.setup()
    const place = { address: 'Мясницкая улица, 20, Москва', lat: 55.76, lon: 37.63 }
    const routes = {
      'GET /api/tasks/today': () => jsonResponse(demoTasks),
      'GET /api/features': () => jsonResponse(demoFeatures),
      'GET /api/for-you': () => jsonResponse([]),
      'GET /api/profile': () =>
        jsonResponse({ name: '', birthDate: '', address: '', lat: 0, lon: 0, hasLocation: false, contactName: '', contactPhone: '', health: '' }),
      ['GET /api/geo/search?q=' + encodeURIComponent('Мясницкая 20')]: () => jsonResponse([place]),
      'PUT /api/profile/address': () => {
        saved = { ...place, isDefault: false }
        return jsonResponse(saved)
      },
      'GET /api/pharmacies/nearby': () => jsonResponse({ location: saved, pharmacies: [pharmacy] }),
    }
    let saved = { address: 'Москва, центр', lat: 55.75, lon: 37.62, isDefault: true }
    mockFetch(routes)
    render(<App />)

    await user.click(screen.getByLabelText('Профиль'))
    await user.type(await screen.findByLabelText('Поиск адреса'), 'Мясницкая 20{Enter}')
    await user.click(await screen.findByRole('button', { name: place.address }))
    await screen.findByText('Адрес сохранён')
    await user.click(screen.getByLabelText('Закрыть'))

    await user.click(await screen.findByRole('button', { name: 'Аптеки рядом' }))
    expect(await screen.findByText('Рядом с: ' + place.address)).toBeInTheDocument()
  })

  it('плитки «Запись к врачу» и «Соцпомощь» открывают свои разделы', async () => {
    const user = userEvent.setup()
    const features = [
      ...demoFeatures,
      { id: 'doctor', title: 'Запись к врачу', icon: 'stethoscope', color: 'rose' },
      { id: 'social', title: 'Соцпомощь', icon: 'heart', color: 'orange' },
    ]
    const clinic = { id: 2, kind: 'clinic', name: 'ГП № 2', address: '', lat: 55.76, lon: 37.64, rating: 4, reviews: 5, distanceKm: 0.8 }
    mockFetch({
      'GET /api/tasks/today': () => jsonResponse(demoTasks),
      'GET /api/features': () => jsonResponse(features),
      'GET /api/for-you': () => jsonResponse([]),
      'GET /api/clinics': () =>
        jsonResponse({ location: location, hasRegistration: true, regAddress: 'x', myClinic: clinic, myClinicChosen: false, nearby: [] }),
      'GET /api/clinics/2': () => jsonResponse({ clinic: clinic, doctors: [], specialties: [] }),
      'GET /api/social/nearby': () => jsonResponse({ location: location, places: [] }),
    })
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'Запись к врачу' }))
    await user.click(await screen.findByRole('button', { name: /ГП № 2/ }))
    expect(await screen.findByText('Врачей не нашли')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))
    expect(await screen.findByText('Ваша поликлиника · по прописке')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))

    await user.click(await screen.findByRole('button', { name: 'Соцпомощь' }))
    expect(await screen.findByText('Рядом организаций соцпомощи не нашли')).toBeInTheDocument()

    await user.click(within(screen.getByRole('navigation')).getByText('Дом'))
    expect(await screen.findByText('Задачи на сегодня')).toBeInTheDocument()
  })

  it('нижнее меню: документы → льгота → назад, помощь → инструкция; из «Для вас» — в замену паспорта', async () => {
    const user = userEvent.setup()
    const benefitsList = {
      categories: ['Документы'],
      summary: { fit: 1, inProgress: 0, approved: 0 },
      items: [{ id: 'passport', category: 'Документы', title: 'Замена паспорта', short: 'За 90 дней', auto: false, fit: true, fitReason: 'Скоро 45', status: 'not_started', stale: false, docsDone: 0, docsTotal: 1 }],
    }
    const passport = {
      benefit: { id: 'passport', category: 'Документы', title: 'Замена паспорта', short: 'За 90 дней', who: 'В 20 и 45 лет', documents: ['Паспорт'], steps: ['Подайте заявление'], links: [{ title: 'Госуслуги', url: 'https://www.gosuslugi.ru/x' }], auto: false },
      fit: true, fitReason: 'Скоро 45', state: { status: 'not_started', docs: [] }, stale: false,
    }
    mockFetch({
      'GET /api/tasks/today': () => jsonResponse(demoTasks),
      'GET /api/features': () => jsonResponse(demoFeatures),
      'GET /api/for-you': () => jsonResponse([{ type: 'passport', title: 'Скоро менять паспорт', text: 'Через месяц 45' }]),
      'GET /api/benefits': () => jsonResponse(benefitsList),
      'GET /api/benefits/passport': () => jsonResponse(passport),
      'GET /api/guides': () => jsonResponse([{ id: 'call', title: 'Как позвонить', short: 'Номер', icon: 'phone', important: false, read: false }]),
      'GET /api/guides/call': () => jsonResponse({ guide: { id: 'call', title: 'Как позвонить', icon: 'phone', steps: ['Наберите номер'] }, read: false }),
    })
    render(<App />)
    const nav = screen.getByRole('navigation')

    await user.click(within(nav).getByText('Документы'))
    await user.click(await screen.findByRole('button', { name: /Замена паспорта/ }))
    expect(await screen.findByText('В 20 и 45 лет')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))
    expect(await screen.findByText('Документы и льготы')).toBeInTheDocument()

    await user.click(within(nav).getByText('Помощь'))
    await user.click(await screen.findByRole('button', { name: /Как позвонить/ }))
    expect(await screen.findByText('Наберите номер')).toBeInTheDocument()
    expect(within(nav).getByText('Помощь').closest('button')).toHaveAttribute('aria-current', 'page')

    await user.click(within(nav).getByText('Дом'))
    await user.click(await screen.findByRole('button', { name: /Для вас/ }))
    await user.click(screen.getByRole('button', { name: 'Как заменить паспорт' }))
    expect(await screen.findByText('В 20 и 45 лет')).toBeInTheDocument()
    expect(within(nav).getByText('Документы').closest('button')).toHaveAttribute('aria-current', 'page')
    localStorage.removeItem('forYouOpen')
  })

  it('кнопка из бота открывает миниапп сразу на лекарстве, название берётся с сервера', async () => {
    window.WebApp = { initData: 'signed', initDataUnsafe: { start_param: 'med_ibuprofen' } }
    const fetchMock = mockAll({})
    render(<App />)

    expect(await screen.findByText('70 ₽')).toBeInTheDocument()
    expect(screen.getByText('Ибупрофен')).toBeInTheDocument()
    expect(screen.getByText('таблетки 200 мг, 20 шт')).toBeInTheDocument()
    const call = fetchMock.mock.calls.find((c) => c[0] === '/api/medicines/ibuprofen/offers')
    expect(call[1].headers['X-Max-Init-Data']).toBe('signed')
    delete window.WebApp
  })
})
