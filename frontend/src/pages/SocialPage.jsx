import { useEffect, useState } from 'react'
import { getSocialNearby } from '../api'
import CallSheet from '../components/CallSheet'
import Icon from '../components/Icon'
import LocationRow from '../components/LocationRow'
import { formatDistance, formatRating, reviewsText } from '../format'

function SocialPage({ onBack, onOpenProfile }) {
  const [location, setLocation] = useState(null)
  const [places, setPlaces] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [calling, setCalling] = useState(null)

  async function loadPlaces() {
    setLoading(true)
    setError('')
    try {
      const data = await getSocialNearby()
      setLocation(data.location)
      setPlaces(data.places || [])
    } catch (err) {
      setError(err.message)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadPlaces()
  }, [])

  let list
  if (loading) {
    list = <p className="card card-note">Загрузка…</p>
  } else if (error !== '') {
    list = (
      <div className="card card-note">
        <p>Не удалось загрузить организации</p>
        <button type="button" className="link-btn" onClick={loadPlaces}>
          Повторить
        </button>
      </div>
    )
  } else if (places.length === 0) {
    list = <p className="card card-note">Рядом организаций соцпомощи не нашли</p>
  } else {
    list = (
      <ul className="pharmacy-list">
        {places.map((p) => (
          <li key={p.id} className="card social">
            <button
              type="button"
              className="social-btn"
              onClick={() => setCalling(p)}
              disabled={!p.phone}
              aria-label={p.phone ? 'Позвонить: ' + p.name : p.name + ', телефон не указан'}
            >
              <span className="social-main">
                <span className="pharmacy-name">{p.name}</span>
                <span className="pharmacy-rating">
                  <span className="star">★</span> {formatRating(p.rating)}
                  <span className="pharmacy-reviews"> · {reviewsText(p.reviews)}</span>
                </span>
                <span className="pharmacy-address">
                  {p.address ? p.address + ' · ' : ''}
                  {formatDistance(p.distanceKm)}
                </span>
                <span className={p.phone ? 'social-phone' : 'social-phone social-phone-none'}>
                  {p.phone || 'Телефон не указан'}
                </span>
              </span>
              {p.phone && (
                <span className="social-call">
                  <Icon name="phone" size={22} />
                </span>
              )}
            </button>
          </li>
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
      <h1 className="page-title">Соцпомощь</h1>
      <p className="page-hint">Центры социального обслуживания и службы помощи рядом с вами</p>
      {location && <LocationRow location={location} onOpenProfile={onOpenProfile} />}
      <div className="section-gap">{list}</div>
      {calling && <CallSheet place={calling} onClose={() => setCalling(null)} />}
    </div>
  )
}

export default SocialPage
