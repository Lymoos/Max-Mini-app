import { useState } from 'react'
import { saveRegistration } from '../api'
import AddressPicker from './AddressPicker'

function RegistrationSheet({ home, onSaved, onClose }) {
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function save(place) {
    setSaving(true)
    setError('')
    try {
      await saveRegistration({ address: place.address, lat: place.lat, lon: place.lon })
      onSaved()
    } catch (err) {
      setError(err.message)
      setSaving(false)
    }
  }

  return (
    <div className="overlay overlay-top" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Адрес регистрации">
        <h2 className="sheet-title">Адрес регистрации</h2>
        <p className="sheet-hint">
          Обычно к поликлинике прикрепляют по адресу прописки. По нему покажем вашу поликлинику первой.
        </p>

        {home && !home.isDefault && (
          <button type="button" className="btn btn-secondary same-address" onClick={() => save(home)} disabled={saving}>
            Совпадает с адресом проживания
            <span className="same-address-text">{home.address}</span>
          </button>
        )}

        <AddressPicker address="" lat={0} lon={0} hasLocation={false} onPick={save} onClear={() => {}} searchLabel="Поиск адреса прописки" />

        {error !== '' && <p className="error-text">{error}</p>}
        <button type="button" className="btn btn-secondary sheet-skip" onClick={onClose}>
          Пропустить
        </button>
      </div>
    </div>
  )
}

export default RegistrationSheet
