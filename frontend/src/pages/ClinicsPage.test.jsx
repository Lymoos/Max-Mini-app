import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import ClinicsPage from './ClinicsPage'

vi.mock('../components/MapPicker', () => ({ default: () => <div>карта</div> }))

const home = { address: 'Мясницкая улица, 20, Москва', lat: 55.76, lon: 37.63, isDefault: false }
const gp = { id: 2, kind: 'clinic', name: 'Городская поликлиника № 2', address: 'Покровка, 1', lat: 55.76, lon: 37.64, rating: 3.9, reviews: 50, distanceKm: 0.8, isState: true }
const medsi = { id: 3, kind: 'clinic', name: 'Медси', address: '', lat: 55.75, lon: 37.62, rating: 4.8, reviews: 300, distanceKm: 0.3, isState: false }

function clinics(extra) {
  return { location: home, hasRegistration: false, regAddress: '', myClinic: null, myClinicChosen: false, nearby: [medsi, gp], ...extra }
}

function puts(fetchMock, path) {
  return fetchMock.mock.calls.filter((c) => c[0] === path && c[1] && c[1].method === 'PUT').map((c) => JSON.parse(c[1].body))
}

describe('ClinicsPage', () => {
  it('в первый раз просит прописку; «совпадает с адресом проживания» сохраняет домашний адрес', async () => {
    const user = userEvent.setup()
    let answer = clinics()
    const fetchMock = mockFetch({
      'GET /api/clinics': () => jsonResponse(answer),
      'PUT /api/profile/registration': () => {
        answer = clinics({ hasRegistration: true, regAddress: home.address, myClinic: gp, nearby: [medsi] })
        return jsonResponse({ hasRegistration: true })
      },
    })
    render(<ClinicsPage onOpenClinic={() => {}} onBack={() => {}} />)

    const sheet = await screen.findByRole('dialog', { name: 'Адрес регистрации' })
    await user.click(within(sheet).getByRole('button', { name: /Совпадает с адресом проживания/ }))

    expect(puts(fetchMock, '/api/profile/registration')).toEqual([{ address: home.address, lat: 55.76, lon: 37.63 }])
    expect(screen.queryByRole('dialog', { name: 'Адрес регистрации' })).not.toBeInTheDocument()
    expect(await screen.findByText('Ваша поликлиника · по прописке')).toBeInTheDocument()
    expect(screen.getByText('Проверьте в полисе ОМС или на Госуслугах')).toBeInTheDocument()
    expect(screen.getByText('Прописка: ' + home.address, { exact: false })).toBeInTheDocument()
  })

  it('можно пропустить, тогда предлагаем указать прописку позже', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/clinics': () => jsonResponse(clinics()) })
    render(<ClinicsPage onOpenClinic={() => {}} onBack={() => {}} />)

    await user.click(await screen.findByRole('button', { name: 'Пропустить' }))
    expect(screen.getByText('Укажите адрес прописки, чтобы показать вашу поликлинику первой.')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Указать прописку' }))
    expect(screen.getByRole('dialog', { name: 'Адрес регистрации' })).toBeInTheDocument()
  })

  it('при повторном открытии с пропиской окно не показывается; своя поликлиника первая', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/clinics': () => jsonResponse(clinics({ hasRegistration: true, regAddress: 'x', myClinic: gp, nearby: [medsi] })) })
    const onOpenClinic = vi.fn()
    render(<ClinicsPage onOpenClinic={onOpenClinic} onBack={() => {}} />)

    await user.click(await screen.findByRole('button', { name: /Городская поликлиника № 2/ }))
    expect(onOpenClinic).toHaveBeenCalledWith(2)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Медси/ }))
    expect(onOpenClinic).toHaveBeenCalledWith(3)
  })

  it('«Это не моя» → выбрать свою → сбросить на автоматическую', async () => {
    const user = userEvent.setup()
    let answer = clinics({ hasRegistration: true, regAddress: 'x', myClinic: gp, nearby: [medsi] })
    const fetchMock = mockFetch({
      'GET /api/clinics': () => jsonResponse(answer),
      'PUT /api/profile/clinic': (options) => {
        const id = JSON.parse(options.body).clinicId
        answer = id
          ? clinics({ hasRegistration: true, regAddress: 'x', myClinic: medsi, myClinicChosen: true, nearby: [gp] })
          : clinics({ hasRegistration: true, regAddress: 'x', myClinic: gp, nearby: [medsi] })
        return jsonResponse({ clinicId: id })
      },
    })
    render(<ClinicsPage onOpenClinic={() => {}} onBack={() => {}} />)

    await user.click(await screen.findByRole('button', { name: 'Это не моя' }))
    expect(screen.getByText('Выберите свою поликлинику')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Моя' }))

    expect(await screen.findByText('Ваша поликлиника')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Медси/ })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Определять по прописке' }))
    expect(await screen.findByText('Ваша поликлиника · по прописке')).toBeInTheDocument()
    expect(puts(fetchMock, '/api/profile/clinic')).toEqual([{ clinicId: 3 }, { clinicId: 0 }])
  })

  it('рядом с пропиской нет районной поликлиники', async () => {
    mockFetch({ 'GET /api/clinics': () => jsonResponse(clinics({ hasRegistration: true, regAddress: 'Село' })) })
    render(<ClinicsPage onOpenClinic={() => {}} onBack={() => {}} />)
    expect(await screen.findByText(/не нашли районную поликлинику/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Выбрать свою' })).toBeInTheDocument()
  })

  it('ошибка и повтор, назад', async () => {
    const user = userEvent.setup()
    mockFetch({})
    const onBack = vi.fn()
    render(<ClinicsPage onOpenClinic={() => {}} onBack={onBack} />)
    expect(await screen.findByText('Не удалось загрузить поликлиники')).toBeInTheDocument()

    mockFetch({ 'GET /api/clinics': () => jsonResponse(clinics({ hasRegistration: true, nearby: [] })) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Рядом поликлиник не нашли')).toBeInTheDocument()
    await user.click(screen.getByText('Назад'))
    expect(onBack).toHaveBeenCalled()
  })
})
