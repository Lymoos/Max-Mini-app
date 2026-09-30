import { useEffect, useState } from 'react'
import { createTask, getDoctor } from '../api'
import Icon from '../components/Icon'
import { formatDistance, formatRating, formatShortDate, phoneLink, reviewsText, routeLink, yearsText } from '../format'

function todayValue() {
  const d = new Date()
  const month = d.getMonth() + 1
  return d.getFullYear() + '-' + (month < 10 ? '0' + month : month) + '-' + (d.getDate() < 10 ? '0' + d.getDate() : d.getDate())
}

function DoctorPage({ doctorId, onBack }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [showReminder, setShowReminder] = useState(false)
  const [date, setDate] = useState('')
  const [time, setTime] = useState('')
  const [saving, setSaving] = useState(false)
  const [reminderError, setReminderError] = useState('')
  const [reminded, setReminded] = useState('')
  const [minDate] = useState(todayValue)

  async function loadDoctor() {
    setError('')
    try {
      setData(await getDoctor(doctorId))
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadDoctor()
  }, [doctorId])

  async function saveReminder(e) {
    e.preventDefault()
    if (date === '' || time === '' || saving) {
      return
    }
    setSaving(true)
    setReminderError('')
    try {
      await createTask({
        title: 'Приём: ' + data.doctor.specialty.toLowerCase() + ', ' + data.clinic.name,
        time: time,
        kind: 'doctor',
        date: date,
      })
      setReminded(formatShortDate(date) + ' в ' + time)
      setShowReminder(false)
    } catch (err) {
      setReminderError(err.message)
    }
    setSaving(false)
  }

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить врача</p>
        <button type="button" className="link-btn" onClick={loadDoctor}>
          Повторить
        </button>
      </div>
    )
  } else if (data === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const { doctor, clinic, booking } = data
    content = (
      <>
        <div className="card doctor-card">
          <span className="doctor-icon doctor-icon-big">
            <Icon name="stethoscope" size={28} />
          </span>
          <h1 className="doctor-title">{doctor.name}</h1>
          <p className="doctor-specialty">{doctor.specialty}</p>
          <p className="doctor-info">
            Стаж {yearsText(doctor.experience)}
            {doctor.category ? ' · ' + doctor.category : ''}
          </p>
          <p className="pharmacy-rating">
            <span className="star">★</span> {formatRating(doctor.rating)}
            <span className="pharmacy-reviews"> · {reviewsText(doctor.reviews)}</span>
          </p>
        </div>

        <div className="card shop-info">
          <p className="pharmacy-name">{clinic.name}</p>
          <p className="pharmacy-address">
            {clinic.address ? clinic.address + ' · ' : ''}
            {formatDistance(clinic.distanceKm)}
          </p>
          <a className="route-link" href={routeLink(clinic.lat, clinic.lon)} target="_blank" rel="noreferrer">
            Построить маршрут
          </a>
        </div>

        <h2 className="block-title promos-title">Как записаться</h2>
        <div className="card booking">
          <p className="booking-text">
            Запись идёт на официальном портале: войдите через Госуслуги, выберите поликлинику и врача.
          </p>
          <a className="btn btn-primary booking-btn" href={booking.url} target="_blank" rel="noreferrer">
            {booking.title}
          </a>
          <a className="btn btn-secondary booking-btn" href={phoneLink(booking.phone)}>
            <Icon name="phone" size={20} />
            Позвонить {booking.phone}
          </a>
        </div>

        {reminded !== '' ? (
          <p className="saved-note reminder-done">Напомним о приёме {reminded}</p>
        ) : (
          <button type="button" className="btn btn-secondary reminder-btn" onClick={() => setShowReminder(true)}>
            Я записался — напомнить
          </button>
        )}
        <p className="demo-note">Данные врача — пример для демонстрации</p>
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
      {showReminder && (
        <div className="overlay overlay-top" onClick={() => setShowReminder(false)}>
          <form
            className="sheet"
            onSubmit={saveReminder}
            onClick={(e) => e.stopPropagation()}
            role="dialog"
            aria-label="Напомнить о приёме"
          >
            <h2 className="sheet-title">Когда приём?</h2>
            <label className="field">
              <span className="field-label">Дата</span>
              <input className="field-input" type="date" min={minDate} value={date} onChange={(e) => setDate(e.target.value)} />
            </label>
            <label className="field">
              <span className="field-label">Время</span>
              <input className="field-input" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
            </label>
            {reminderError !== '' && <p className="error-text">{reminderError}</p>}
            <div className="sheet-buttons">
              <button type="button" className="btn btn-secondary" onClick={() => setShowReminder(false)}>
                Отмена
              </button>
              <button type="submit" className="btn btn-primary" disabled={date === '' || time === '' || saving}>
                Напомнить
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}

export default DoctorPage
