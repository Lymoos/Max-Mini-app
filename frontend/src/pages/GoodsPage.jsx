import { useEffect, useRef, useState } from 'react'
import { getNearbyShops, getProductOffers, suggestProducts } from '../api'
import CatalogSuggestions from '../components/CatalogSuggestions'
import Icon from '../components/Icon'
import LocationRow from '../components/LocationRow'
import PlaceCard from '../components/PlaceCard'

function GoodsPage({ product, onPickProduct, onClearProduct, onOpenShop, onOpenProfile, onBack }) {
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
      if (product) {
        data = await getProductOffers(product.id)
        list = data.offers
      } else {
        data = await getNearbyShops()
        list = data.shops
      }
      if (requestId !== lastRequest.current) {
        return
      }
      setLocation(data.location)
      setInfo(data.product || null)
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
  }, [product])

  function handlePick(p) {
    setQuery('')
    setShowHints(false)
    setSearchError('')
    onPickProduct(p)
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
      const found = await suggestProducts(q)
      if (found.length === 0) {
        setSearchError('Такой товар не нашли. Попробуйте написать по-другому')
      } else {
        handlePick(found[0])
      }
    } catch (err) {
      setSearchError(err.message)
    }
  }

  // из бота приходит только id — название берём из ответа сервера
  const shown = info && product && info.id === product.id ? info : product

  let list
  if (loading) {
    list = <p className="card card-note">Загрузка…</p>
  } else if (error !== '') {
    list = (
      <div className="card card-note">
        <p>Не удалось загрузить магазины</p>
        <button type="button" className="link-btn" onClick={loadList}>
          Повторить
        </button>
      </div>
    )
  } else if (items.length === 0) {
    list = (
      <p className="card card-note">{product ? 'Рядом нет магазинов, где есть этот товар' : 'Рядом магазинов не нашли'}</p>
    )
  } else {
    list = (
      <ul className="pharmacy-list">
        {items.map((shop) => (
          <PlaceCard key={shop.id} place={shop} onOpen={onOpenShop} />
        ))}
      </ul>
    )
  }

  return (
    <div className="page">
      <button type="button" className="back-btn" onClick={onBack}>
        <Icon name="back" size={20} />
        Назад
      </button>
      <h1 className="page-title">Товары рядом</h1>

      <div className="med-search">
        <form className="search-box" onSubmit={handleSubmit} role="search">
          <span className="search-icon">
            <Icon name="search" />
          </span>
          <input
            className="search-input"
            type="text"
            placeholder="Название товара"
            aria-label="Название товара"
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
            fetcher={suggestProducts}
            icon="cart"
            subtitle="Найти в магазинах рядом"
            onPick={handlePick}
          />
        )}
      </div>
      {searchError !== '' && <p className="error-text">{searchError}</p>}

      {location && <LocationRow location={location} onOpenProfile={onOpenProfile} />}

      {product ? (
        <div className="card med-chosen">
          <span className="med-chosen-icon goods-icon">
            <Icon name="cart" />
          </span>
          <div className="med-chosen-text">
            <p className="med-chosen-name">{shown.name}</p>
            <p className="med-chosen-form">{shown.unit}</p>
          </div>
          <button type="button" className="address-remove" onClick={onClearProduct} aria-label="Показать все магазины">
            <Icon name="close" size={18} />
          </button>
        </div>
      ) : (
        <div className="med-list-header">
          <h2 className="block-title">Лучшие магазины рядом</h2>
          <p className="med-list-hint">Нажмите на магазин, чтобы увидеть акции</p>
        </div>
      )}

      {list}
    </div>
  )
}

export default GoodsPage
