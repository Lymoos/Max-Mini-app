import { useEffect, useState } from 'react'
import { getShop } from '../api'
import Icon from '../components/Icon'
import { formatDistance, formatRating, formatShortDate, reviewsText, routeLink } from '../format'

function ShopPage({ shopId, onBack }) {
  const [shop, setShop] = useState(null)
  const [promos, setPromos] = useState([])
  const [error, setError] = useState('')

  async function loadShop() {
    setError('')
    try {
      const data = await getShop(shopId)
      setShop(data.shop)
      setPromos(data.promos || [])
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadShop()
  }, [shopId])

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить магазин</p>
        <button type="button" className="link-btn" onClick={loadShop}>
          Повторить
        </button>
      </div>
    )
  } else if (shop === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    content = (
      <>
        <h1 className="page-title shop-title">{shop.name}</h1>
        <div className="card shop-info">
          <p className="pharmacy-rating">
            <span className="star">★</span> {formatRating(shop.rating)}
            <span className="pharmacy-reviews"> · {reviewsText(shop.reviews)}</span>
          </p>
          <p className="pharmacy-address">
            {shop.address ? shop.address + ' · ' : ''}
            {formatDistance(shop.distanceKm)}
          </p>
          {shop.birthdayDiscount > 0 && (
            <p className="shop-birthday">
              <Icon name="gift" size={18} />
              Скидка {shop.birthdayDiscount}% в день рождения
            </p>
          )}
          <a className="route-link" href={routeLink(shop.lat, shop.lon)} target="_blank" rel="noreferrer">
            Построить маршрут
          </a>
        </div>

        <h2 className="block-title promos-title">Акции сейчас</h2>
        {promos.length === 0 ? (
          <p className="card card-note">Сейчас акций нет</p>
        ) : (
          <ul className="promo-list">
            {promos.map((p) => (
              <li key={p.productId} className="card promo">
                <div className="promo-main">
                  <p className="promo-name">{p.name}</p>
                  <p className="promo-unit">
                    {p.unit} · до {formatShortDate(p.promoUntil)}
                  </p>
                </div>
                <div className="promo-side">
                  <span className="promo-discount">−{p.discount}%</span>
                  <p className="pharmacy-price price-promo">{p.price} ₽</p>
                  <p className="old-price">{p.oldPrice} ₽</p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </>
    )
  }

  return (
    <div className="page">
      <button type="button" className="back-btn" onClick={onBack}>
        <Icon name="back" size={20} />
        Назад
      </button>
      {content}
    </div>
  )
}

export default ShopPage
