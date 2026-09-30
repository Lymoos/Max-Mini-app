import { useEffect, useRef, useState } from 'react'
import { getNearbyPharmacies, getOffers, suggestMedicines } from '../api'
import CatalogSuggestions from '../components/CatalogSuggestions'
import Icon from '../components/Icon'
import LocationRow from '../components/LocationRow'
import PlaceCard from '../components/PlaceCard'

function MedicinesPage({ medicine, onPickMedicine, onClearMedicine, onOpenProfile }) {
  const [query, setQuery] = useState('')
  const [showHints, setShowHints] = useState(false)
  const [searchError, setSearchError] = useState('')
  const [location, setLocation] = useState(null)
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [info, setInfo] = useState(null)
  const lastRequest = useRef(0)

  async function loadList() {
    const requestId = lastRequest.current + 1
    lastRequest.current = requestId
    setLoading(true)
    setError('')
    try {
      let data
      let list
      if (medicine) {
        data = await getOffers(medicine.id)
        list = data.offers
      } else {
        data = await getNearbyPharmacies()
        list = data.pharmacies
      }
      if (requestId !== lastRequest.current) {
        return
      }
      setLocation(data.location)
      setInfo(data.medicine || null)
      setItems(list || [])
    } catch (err) {
      if (requestId !== lastRequest.current) {
        return
      }
      setError(err.message)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadList()
  }, [medicine])

  function handlePick(m) {
    setQuery('')
    setShowHints(false)
    setSearchError('')
    onPickMedicine(m)
  }

  async function handleSubmit(e) {
    e.preventDefault()
    const q = query.trim()
    if (q === '') {
      return
    }
    setShowHints(false)
    setSearchError('')
    try {
      const found = await suggestMedicines(q)
      if (found.length === 0) {
        setSearchError('Такое лекарство не нашли. Проверьте название')
      } else {
        handlePick(found[0])
      }
    } catch (err) {
      setSearchError(err.message)
    }
  }

  // из бота приходит только id — название берём из ответа сервера
  const shown = info && medicine && info.id === medicine.id ? info : medicine

  let list
  if (loading) {
    list = <p className="card card-note">Загрузка…</p>
  } else if (error !== '') {
    list = (
      <div className="card card-note">
        <p>Не удалось загрузить аптеки</p>
        <button type="button" className="link-btn" onClick={loadList}>
          Повторить
        </button>
      </div>
    )
  } else if (items.length === 0) {
    list = (
      <p className="card card-note">
        {medicine ? 'Рядом нет аптек, где есть это лекарство' : 'Рядом аптек не нашли'}
      </p>
    )
  } else {
    list = (
      <ul className="pharmacy-list">
        {items.map((p) => (
          <PlaceCard key={p.id} place={p} />
        ))}
      </ul>
    )
  }

  return (
    <div className="page">
      <h1 className="page-title">Лекарства</h1>

      <div className="med-search">
        <form className="search-box" onSubmit={handleSubmit} role="search">
          <span className="search-icon">
            <Icon name="search" />
          </span>
          <input
            className="search-input"
            type="text"
            placeholder="Название лекарства"
            aria-label="Название лекарства"
            enterKeyHint="search"
            maxLength={100}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setShowHints(true)
              setSearchError('')
            }}
          />
        </form>
        {showHints && (
          <CatalogSuggestions
            query={query}
            fetcher={suggestMedicines}
            icon="pill"
            subtitle="Найти в ближайших аптеках"
            onPick={handlePick}
          />
        )}
      </div>
      {searchError !== '' && <p className="error-text">{searchError}</p>}

      {location && <LocationRow location={location} onOpenProfile={onOpenProfile} />}

      {medicine ? (
        <div className="card med-chosen">
          <span className="med-chosen-icon">
            <Icon name="pill" />
          </span>
          <div className="med-chosen-text">
            <p className="med-chosen-name">{shown.name}</p>
            <p className="med-chosen-form">{shown.form}</p>
          </div>
          <button type="button" className="address-remove" onClick={onClearMedicine} aria-label="Показать все аптеки">
            <Icon name="close" size={18} />
          </button>
        </div>
      ) : (
        <div className="med-list-header">
          <h2 className="block-title">Лучшие аптеки рядом</h2>
          <p className="med-list-hint">По рейтингу и расстоянию</p>
        </div>
      )}

      {list}
    </div>
  )
}

export default MedicinesPage
