import { useEffect, useRef, useState } from 'react'
import { loadYandexMaps } from '../yandexMaps'

const moscow = [55.7539, 37.6208]

function MapPicker({ lat, lon, onPick }) {
  const box = useRef(null)
  const map = useRef(null)
  const mark = useRef(null)
  const pickRef = useRef(onPick)
  const [ready, setReady] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    pickRef.current = onPick
  }, [onPick])

  useEffect(() => {
    let closed = false
    loadYandexMaps()
      .then((ymaps) => {
        if (closed) {
          return
        }
        map.current = new ymaps.Map(box.current, {
          center: lat || lon ? [lat, lon] : moscow,
          zoom: 15,
          controls: ['zoomControl'],
        })
        map.current.events.add('click', (e) => {
          const coords = e.get('coords')
          pickRef.current(coords[0], coords[1])
        })
        setReady(true)
      })
      .catch((err) => setError(err.message))

    return () => {
      closed = true
      if (map.current) {
        map.current.destroy()
        map.current = null
      }
    }
  }, [])

  useEffect(() => {
    if (!ready) {
      return
    }
    if (mark.current) {
      map.current.geoObjects.remove(mark.current)
      mark.current = null
    }
    if (lat || lon) {
      mark.current = new window.ymaps.Placemark([lat, lon], {}, { preset: 'islands#blueCircleDotIcon' })
      map.current.geoObjects.add(mark.current)
      map.current.setCenter([lat, lon])
    }
  }, [ready, lat, lon])

  return (
    <div className="map">
      {error !== '' ? <p className="error-text">{error}</p> : <div className="map-view" ref={box} />}
      <p className="map-hint">Нажмите на карту, чтобы выбрать место</p>
    </div>
  )
}

export default MapPicker
