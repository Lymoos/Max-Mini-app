import { formatDistance, formatRating, formatShortDate, reviewsText, routeLink } from '../format'

function PlaceCard({ place, onOpen }) {
  const main = (
    <>
      <p className="pharmacy-name">{place.name}</p>
      <p className="pharmacy-rating">
        <span className="star">★</span> {formatRating(place.rating)}
        <span className="pharmacy-reviews"> · {reviewsText(place.reviews)}</span>
      </p>
      <p className="pharmacy-address">
        {place.address ? place.address + ' · ' : ''}
        {formatDistance(place.distanceKm)}
      </p>
      {place.promoUntil && <p className="promo-until">Акция до {formatShortDate(place.promoUntil)}</p>}
    </>
  )

  return (
    <li className="card pharmacy">
      {onOpen ? (
        <button type="button" className="pharmacy-main pharmacy-open" onClick={() => onOpen(place)}>
          {main}
        </button>
      ) : (
        <div className="pharmacy-main">{main}</div>
      )}
      <div className="pharmacy-side">
        {place.price !== undefined && (
          <div className="price-box">
            <p className={place.oldPrice ? 'pharmacy-price price-promo' : 'pharmacy-price'}>{place.price} ₽</p>
            {place.oldPrice > 0 && <p className="old-price">{place.oldPrice} ₽</p>}
          </div>
        )}
        {place.cheapest && <span className="badge">Дешевле всего</span>}
        <a className="route-link" href={routeLink(place.lat, place.lon)} target="_blank" rel="noreferrer">
          Маршрут
        </a>
      </div>
    </li>
  )
}

export default PlaceCard
