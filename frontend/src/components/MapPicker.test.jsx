import { act, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MapPicker from './MapPicker'
import { loadYandexMaps } from '../yandexMaps'

vi.mock('../yandexMaps', () => ({ loadYandexMaps: vi.fn() }))

let created

function fakeYmaps() {
  created = { maps: [], placemarks: [] }
  class FakeMap {
    constructor(el, options) {
      this.el = el
      this.options = options
      this.handlers = {}
      this.objects = []
      this.center = options.center
      this.destroyed = false
      this.events = { add: (name, fn) => (this.handlers[name] = fn) }
      this.geoObjects = {
        add: (o) => this.objects.push(o),
        remove: (o) => (this.objects = this.objects.filter((x) => x !== o)),
      }
      created.maps.push(this)
    }
    setCenter(c) {
      this.center = c
    }
    destroy() {
      this.destroyed = true
    }
  }
  class FakePlacemark {
    constructor(coords) {
      this.coords = coords
      created.placemarks.push(this)
    }
  }
  const ymaps = { Map: FakeMap, Placemark: FakePlacemark }
  window.ymaps = ymaps
  return ymaps
}

describe('MapPicker', () => {
  beforeEach(() => {
    delete window.ymaps
  })

  it('создаёт карту, ставит метку и отдаёт координаты клика', async () => {
    loadYandexMaps.mockResolvedValue(fakeYmaps())
    const onPick = vi.fn()
    const { rerender, unmount } = render(<MapPicker lat={55.75} lon={37.62} onPick={onPick} />)
    await act(async () => {})

    const map = created.maps[0]
    expect(map.options.center).toEqual([55.75, 37.62])
    expect(map.objects).toHaveLength(1)
    expect(map.objects[0].coords).toEqual([55.75, 37.62])

    map.handlers.click({ get: () => [55.8, 37.7] })
    expect(onPick).toHaveBeenCalledWith(55.8, 37.7)

    rerender(<MapPicker lat={55.8} lon={37.7} onPick={onPick} />)
    expect(map.objects).toHaveLength(1)
    expect(map.objects[0].coords).toEqual([55.8, 37.7])
    expect(map.center).toEqual([55.8, 37.7])

    unmount()
    expect(map.destroyed).toBe(true)
  })

  it('без точки центрирует на Москве и не ставит метку', async () => {
    loadYandexMaps.mockResolvedValue(fakeYmaps())
    render(<MapPicker lat={0} lon={0} onPick={() => {}} />)
    await act(async () => {})
    expect(created.maps[0].options.center).toEqual([55.7539, 37.6208])
    expect(created.maps[0].objects).toHaveLength(0)
  })

  it('если карта не загрузилась, показывает ошибку', async () => {
    loadYandexMaps.mockRejectedValue(new Error('Не удалось загрузить карту'))
    render(<MapPicker lat={0} lon={0} onPick={() => {}} />)
    expect(await screen.findByText('Не удалось загрузить карту')).toBeInTheDocument()
  })

  it('если окно закрыли до загрузки карты, карта не создаётся', async () => {
    let finish
    loadYandexMaps.mockReturnValue(new Promise((resolve) => (finish = resolve)))
    const { unmount } = render(<MapPicker lat={0} lon={0} onPick={() => {}} />)
    unmount()
    const ymaps = fakeYmaps()
    await act(async () => finish(ymaps))
    expect(created.maps).toHaveLength(0)
  })
})
