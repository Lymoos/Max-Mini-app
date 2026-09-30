import { useEffect, useState } from 'react'
import { getProfile, saveAddress, saveProfile, saveRegistration } from '../api'
import AddressPicker from '../components/AddressPicker'
import BirthdayPicker from '../components/BirthdayPicker'
import Icon from '../components/Icon'
import { ageFrom, formatFullDate, yearsText } from '../format'

function ProfilePage({ onClose, onSaved }) {
  const [form, setForm] = useState(null)
  const [saved, setSaved] = useState('')
  const [loadError, setLoadError] = useState('')
  const [error, setError] = useState('')
  const [addressNote, setAddressNote] = useState('')
  const [regNote, setRegNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [showBirthday, setShowBirthday] = useState(false)
  const [showWhy, setShowWhy] = useState(false)
  const [today] = useState(() => new Date())

  async function loadProfile() {
    setLoadError('')
    try {
      const p = await getProfile()
      setForm(p)
      setSaved(JSON.stringify(p))
    } catch (err) {
      setLoadError(err.message)
    }
  }

  useEffect(() => {
    loadProfile()
  }, [])

  function update(field, value) {
    setForm({ ...form, [field]: value })
  }

  async function changeAddress(place) {
    const next = place
      ? { ...form, address: place.address, lat: place.lat, lon: place.lon, hasLocation: true }
      : { ...form, address: '', lat: 0, lon: 0, hasLocation: false }
    setForm(next)
    setAddressNote('')
    setError('')
    try {
      await saveAddress(place ? { address: place.address, lat: place.lat, lon: place.lon } : { address: '' })
      setSaved(JSON.stringify({ ...JSON.parse(saved), address: next.address, lat: next.lat, lon: next.lon, hasLocation: next.hasLocation }))
      setAddressNote(place ? 'Адрес сохранён' : 'Адрес удалён')
      onSaved()
    } catch (err) {
      setError(err.message)
    }
  }

  async function changeRegistration(place) {
    const fields = place
      ? { regAddress: place.address, regLat: place.lat, regLon: place.lon, hasRegistration: true }
      : { regAddress: '', regLat: 0, regLon: 0, hasRegistration: false }
    setForm({ ...form, ...fields })
    setRegNote('')
    setError('')
    try {
      await saveRegistration(place ? { address: place.address, lat: place.lat, lon: place.lon } : { address: '' })
      setSaved(JSON.stringify({ ...JSON.parse(saved), ...fields }))
      setRegNote(place ? 'Прописка сохранена' : 'Прописка удалена')
      onSaved()
    } catch (err) {
      setError(err.message)
    }
  }

  async function save() {
    setSaving(true)
    setError('')
    try {
      await saveProfile(form)
      onSaved()
      onClose()
    } catch (err) {
      setError(err.message)
      setSaving(false)
    }
  }

  function handleSubmit(e) {
    e.preventDefault()
    if (!saving) {
      save()
    }
  }

  // если что-то поменяли и закрыли крестиком, всё равно сохраняем
  function handleClose() {
    if (form !== null && JSON.stringify(form) !== saved) {
      save()
    } else {
      onClose()
    }
  }

  let content
  if (loadError !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить профиль</p>
        <button type="button" className="link-btn" onClick={loadProfile}>
          Повторить
        </button>
      </div>
    )
  } else if (form === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    content = (
      <form className="profile-form" onSubmit={handleSubmit}>
        <div className="card profile-section">
          <label className="field">
            <span className="field-label">Как к вам обращаться</span>
            <input
              className="field-input"
              type="text"
              maxLength={100}
              value={form.name}
              onChange={(e) => update('name', e.target.value)}
            />
          </label>

          <div className="field">
            <div className="field-label-row">
              <span className="field-label" id="birthday-label">
                Дата рождения
              </span>
              <button
                type="button"
                className="why-btn"
                onClick={() => setShowWhy(!showWhy)}
                aria-label="Зачем указывать дату рождения"
                aria-expanded={showWhy}
              >
                <Icon name="question" size={20} />
              </button>
            </div>
            {showWhy && (
              <p className="why-text">
                Покажем магазины и аптеки со скидкой в ваш день рождения и напомним о документах — например, когда
                пора менять паспорт.
              </p>
            )}
            <button
              type="button"
              className="field-input birthday-btn"
              onClick={() => setShowBirthday(true)}
              aria-labelledby="birthday-label"
            >
              {form.birthDate ? (
                <>
                  {formatFullDate(form.birthDate)}
                  <span className="birthday-age"> · {yearsText(ageFrom(form.birthDate, today))}</span>
                </>
              ) : (
                <span className="birthday-empty">Выбрать дату</span>
              )}
            </button>
          </div>
        </div>

        <div className="card profile-section">
          <h2 className="profile-section-title">Адрес</h2>
          <p className="profile-section-hint">По нему ищем аптеки и магазины рядом</p>
          <AddressPicker
            address={form.address}
            lat={form.lat}
            lon={form.lon}
            hasLocation={form.hasLocation}
            onPick={changeAddress}
            onClear={() => changeAddress(null)}
          />
          {addressNote !== '' && <p className="saved-note">{addressNote}</p>}
        </div>

        <div className="card profile-section">
          <h2 className="profile-section-title">Адрес регистрации</h2>
          <p className="profile-section-hint">По прописке определяем вашу поликлинику для записи к врачу</p>
          {!form.hasRegistration && form.hasLocation && (
            <button
              type="button"
              className="chip-btn same-home"
              onClick={() => changeRegistration({ address: form.address, lat: form.lat, lon: form.lon })}
            >
              Совпадает с адресом проживания
            </button>
          )}
          <AddressPicker
            address={form.regAddress}
            lat={form.regLat}
            lon={form.regLon}
            hasLocation={form.hasRegistration}
            onPick={changeRegistration}
            onClear={() => changeRegistration(null)}
            searchLabel="Поиск адреса прописки"
            removeLabel="Убрать прописку"
          />
          {regNote !== '' && <p className="saved-note">{regNote}</p>}
        </div>

        <div className="card profile-section">
          <h2 className="profile-section-title">Близкий человек</h2>
          <p className="profile-section-hint">Кому позвонить, если понадобится помощь</p>
          <label className="field">
            <span className="field-label">Имя</span>
            <input
              className="field-input"
              type="text"
              maxLength={100}
              value={form.contactName}
              onChange={(e) => update('contactName', e.target.value)}
            />
          </label>
          <label className="field">
            <span className="field-label">Телефон</span>
            <input
              className="field-input"
              type="tel"
              placeholder="+7 900 000-00-00"
              maxLength={20}
              value={form.contactPhone}
              onChange={(e) => update('contactPhone', e.target.value)}
            />
          </label>
        </div>

        <div className="card profile-section">
          <label className="field">
            <span className="profile-section-title">Здоровье</span>
            <span className="profile-section-hint">Аллергии, хронические болезни — поможет с подбором лекарств</span>
            <textarea
              className="field-input field-textarea"
              maxLength={1000}
              value={form.health}
              onChange={(e) => update('health', e.target.value)}
            />
          </label>
        </div>

        {error !== '' && <p className="error-text">{error}</p>}
        <button type="submit" className="btn btn-primary profile-save" disabled={saving}>
          {saving ? 'Сохраняем…' : 'Сохранить'}
        </button>
      </form>
    )
  }

  return (
    <div className="profile" role="dialog" aria-label="Профиль">
      <div className="profile-header">
        <h1 className="profile-title">Профиль</h1>
        <button type="button" className="profile-close" onClick={handleClose} aria-label="Закрыть">
          <Icon name="close" />
        </button>
      </div>
      <div className="profile-body">{content}</div>
      {showBirthday && (
        <BirthdayPicker
          value={form.birthDate}
          today={today}
          onDone={(value) => {
            update('birthDate', value)
            setShowBirthday(false)
          }}
          onClose={() => setShowBirthday(false)}
        />
      )}
    </div>
  )
}

export default ProfilePage
