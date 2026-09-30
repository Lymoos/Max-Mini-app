import { useState } from 'react'
import { reverseAddress, searchAddress } from '../api'
import Icon from './Icon'
import MapPicker from './MapPicker'

function AddressPicker({
  address,
  lat,
  lon,
  hasLocation,
  onPick,
  onClear,
  searchLabel = 'Поиск адреса',
  removeLabel = 'Убрать адрес',
}) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [showMap, setShowMap] = useState(false)

  async function handleSearch() {
    if (busy) {
      return
    }
    if (query.trim().length < 3) {
      setError('Напишите адрес подробнее')
      return
    }

    setBusy(true)
    setError('')
    setResults([])
    try {
      const places = await searchAddress(query.trim())
      if (places.length === 0) {
        setError('Адрес не найден. Попробуйте написать по-другому')
      }
      setResults(places)
    } catch (err) {
      setError(err.message)
    }
    setBusy(false)
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleSearch()
    }
  }

  function pickPlace(place) {
    onPick(place)
    setResults([])
    setQuery('')
    setError('')
  }

  async function pickPoint(pointLat, pointLon) {
    if (busy) {
      return
    }
    setBusy(true)
    setError('')
    try {
      const place = await reverseAddress(pointLat, pointLon)
      pickPlace(place)
    } catch (err) {
      setError(err.message)
    }
    setBusy(false)
  }

  function handleLocate() {
    if (!navigator.geolocation) {
      setError('Телефон не даёт определить местоположение')
      return
    }
    setError('')
    navigator.geolocation.getCurrentPosition(
      (pos) => pickPoint(pos.coords.latitude, pos.coords.longitude),
      () => setError('Не удалось определить местоположение. Разрешите доступ или выберите на карте'),
      { timeout: 10000 },
    )
  }

  return (
    <div className="address">
      {hasLocation && (
        <div className="address-chosen">
          <span className="address-chosen-icon">
            <Icon name="pin" size={20} />
          </span>
          <span className="address-chosen-text">{address}</span>
          <button type="button" className="address-remove" onClick={onClear} aria-label={removeLabel}>
            <Icon name="close" size={18} />
          </button>
        </div>
      )}

      <div className="address-search">
        <input
          className="field-input"
          type="text"
          placeholder={hasLocation ? 'Другой адрес' : 'Улица, дом, город'}
          aria-label={searchLabel}
          maxLength={200}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <button type="button" className="btn btn-primary address-find" onClick={handleSearch} disabled={busy}>
          Найти
        </button>
      </div>

      {results.length > 0 && (
        <ul className="address-results">
          {results.map((place, i) => (
            <li key={i}>
              <button type="button" className="address-result" onClick={() => pickPlace(place)}>
                <Icon name="pin" size={18} />
                <span>{place.address}</span>
              </button>
            </li>
          ))}
        </ul>
      )}

      <div className="address-actions">
        <button type="button" className="chip-btn" onClick={() => setShowMap(!showMap)} aria-pressed={showMap}>
          <Icon name="map" size={18} />
          {showMap ? 'Скрыть карту' : 'Выбрать на карте'}
        </button>
        <button type="button" className="chip-btn" onClick={handleLocate} disabled={busy}>
          <Icon name="locate" size={18} />
          Где я сейчас
        </button>
      </div>

      {busy && <p className="address-note">Ищем…</p>}
      {error !== '' && <p className="error-text">{error}</p>}
      {showMap && <MapPicker lat={hasLocation ? lat : 0} lon={hasLocation ? lon : 0} onPick={pickPoint} />}
    </div>
  )
}

export default AddressPicker
