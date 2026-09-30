import { useEffect, useState } from 'react'
import { getClinic } from '../api'
import Icon from '../components/Icon'
import { formatDistance, formatRating, phoneLink, reviewsText, routeLink, yearsText } from '../format'

function ClinicPage({ clinicId, specialtyId, onOpenDoctor, onBack }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [noSpecialist, setNoSpecialist] = useState(false)

  async function loadClinic() {
    setError('')
    try {
      const result = await getClinic(clinicId)
      if (specialtyId) {
        const doctor = result.doctors.find((d) => d.specialtyId === specialtyId)
        if (doctor) {
          setSpecialty(doctor.specialty)
        } else {
          setNoSpecialist(true)
        }
      }
      setData(result)
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadClinic()
  }, [clinicId])

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить поликлинику</p>
        <button type="button" className="link-btn" onClick={loadClinic}>
          Повторить
        </button>
      </div>
    )
  } else if (data === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const clinic = data.clinic
    const doctors = data.doctors.filter((d) => specialty === '' || d.specialty === specialty)
    content = (
      <>
        <h1 className="page-title shop-title">{clinic.name}</h1>
        <div className="card shop-info">
          <p className="pharmacy-address">
            {clinic.address ? clinic.address + ' · ' : ''}
            {formatDistance(clinic.distanceKm)}
          </p>
          {clinic.phone && (
            <a className="route-link" href={phoneLink(clinic.phone)}>
              {clinic.phone}
            </a>
          )}
          <a className="route-link clinic-route" href={routeLink(clinic.lat, clinic.lon)} target="_blank" rel="noreferrer">
            Построить маршрут
          </a>
        </div>

        <h2 className="block-title promos-title">Врачи</h2>
        <p className="demo-note">Список врачей — пример для демонстрации</p>
        {noSpecialist && (
          <p className="card card-note specialty-hint">В этой поликлинике нет врача нужной специальности. Показаны все врачи.</p>
        )}
        {data.specialties.length > 1 && (
          <div className="chips" role="group" aria-label="Специальность">
            <button
              type="button"
              className={specialty === '' ? 'chip chip-active' : 'chip'}
              onClick={() => setSpecialty('')}
              aria-pressed={specialty === ''}
            >
              Все
            </button>
            {data.specialties.map((s) => (
              <button
                key={s}
                type="button"
                className={specialty === s ? 'chip chip-active' : 'chip'}
                onClick={() => setSpecialty(s)}
                aria-pressed={specialty === s}
              >
                {s}
              </button>
            ))}
          </div>
        )}
        {doctors.length === 0 ? (
          <p className="card card-note">Врачей не нашли</p>
        ) : (
          <ul className="pharmacy-list">
            {doctors.map((d) => (
              <li key={d.id}>
                <button type="button" className="card doctor" onClick={() => onOpenDoctor(d.id)}>
                  <span className="doctor-icon">
                    <Icon name="stethoscope" />
                  </span>
                  <span className="doctor-main">
                    <span className="doctor-name">{d.name}</span>
                    <span className="doctor-specialty">{d.specialty}</span>
                    <span className="doctor-info">
                      Стаж {yearsText(d.experience)}
                      {d.category ? ' · ' + d.category : ''}
                    </span>
                    <span className="pharmacy-rating">
                      <span className="star">★</span> {formatRating(d.rating)}
                      <span className="pharmacy-reviews"> · {reviewsText(d.reviews)}</span>
                    </span>
                  </span>
                </button>
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

export default ClinicPage
