import { useEffect, useState } from 'react'
import { getClinics, saveClinic } from '../api'
import Icon from '../components/Icon'
import PlaceCard from '../components/PlaceCard'
import RegistrationSheet from '../components/RegistrationSheet'
import { formatDistance } from '../format'

function ClinicsPage({ specialtyId, autoOpen, onOpenClinic, onBack }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [askRegistration, setAskRegistration] = useState(false)
  const [choosing, setChoosing] = useState(false)
  const [saveError, setSaveError] = useState('')

  async function loadClinics(firstTime) {
    setError('')
    try {
      const result = await getClinics()
      if (firstTime && autoOpen && result.myClinic) {
        onOpenClinic(result.myClinic.id)
        return
      }
      setData(result)
      if (firstTime && !result.hasRegistration) {
        setAskRegistration(true)
      }
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadClinics(true)
  }, [])

  async function chooseClinic(id) {
    setSaveError('')
    try {
      await saveClinic(id)
      setChoosing(false)
      loadClinics(false)
    } catch (err) {
      setSaveError(err.message)
    }
  }

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить поликлиники</p>
        <button type="button" className="link-btn" onClick={() => loadClinics(false)}>
          Повторить
        </button>
      </div>
    )
  } else if (data === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const my = data.myClinic
    content = (
      <>
        {my ? (
          <div className="card my-clinic">
            <span className="my-clinic-badge">{data.myClinicChosen ? 'Ваша поликлиника' : 'Ваша поликлиника · по прописке'}</span>
            <button type="button" className="my-clinic-open" onClick={() => onOpenClinic(my.id)}>
              <span className="pharmacy-name">{my.name}</span>
              <span className="pharmacy-address">
                {my.address ? my.address + ' · ' : ''}
                {formatDistance(my.distanceKm)}
              </span>
            </button>
            {!data.myClinicChosen && <p className="my-clinic-hint">Проверьте в полисе ОМС или на Госуслугах</p>}
            <div className="my-clinic-actions">
              {data.myClinicChosen ? (
                <button type="button" className="link-btn" onClick={() => chooseClinic(0)}>
                  Определять по прописке
                </button>
              ) : (
                <button type="button" className="link-btn" onClick={() => setChoosing(!choosing)}>
                  {choosing ? 'Отмена' : 'Это не моя'}
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className="card my-clinic">
            <p className="my-clinic-empty">
              {data.hasRegistration
                ? 'Рядом с адресом прописки не нашли районную поликлинику. Выберите свою из списка.'
                : 'Укажите адрес прописки, чтобы показать вашу поликлинику первой.'}
            </p>
            <div className="my-clinic-actions">
              {data.hasRegistration ? (
                <button type="button" className="link-btn" onClick={() => setChoosing(!choosing)}>
                  {choosing ? 'Отмена' : 'Выбрать свою'}
                </button>
              ) : (
                <button type="button" className="link-btn" onClick={() => setAskRegistration(true)}>
                  Указать прописку
                </button>
              )}
            </div>
          </div>
        )}
        {data.hasRegistration && (
          <p className="reg-line">
            Прописка: {data.regAddress}{' '}
            <button type="button" className="link-btn reg-change" onClick={() => setAskRegistration(true)}>
              Изменить
            </button>
          </p>
        )}
        {saveError !== '' && <p className="error-text">{saveError}</p>}

        <div className="med-list-header">
          <h2 className="block-title">{choosing ? 'Выберите свою поликлинику' : 'Поликлиники рядом'}</h2>
          <p className="med-list-hint">
            {data.location.isDefault ? 'Рядом с центром Москвы' : 'Рядом с: ' + data.location.address}
          </p>
        </div>
        {data.nearby.length === 0 ? (
          <p className="card card-note">Рядом поликлиник не нашли</p>
        ) : (
          <ul className="pharmacy-list">
            {data.nearby.map((c) =>
              choosing ? (
                <li key={c.id} className="card choose-clinic">
                  <div className="pharmacy-main">
                    <p className="pharmacy-name">{c.name}</p>
                    <p className="pharmacy-address">{formatDistance(c.distanceKm)}</p>
                  </div>
                  <button type="button" className="btn btn-primary choose-btn" onClick={() => chooseClinic(c.id)}>
                    Моя
                  </button>
                </li>
              ) : (
                <PlaceCard key={c.id} place={c} onOpen={(p) => onOpenClinic(p.id)} />
              ),
            )}
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
      <h1 className="page-title">Запись к врачу</h1>
      {specialtyId && data !== null && (
        <p className="card card-note specialty-hint">Выберите поликлинику — сразу покажем врачей нужной специальности</p>
      )}
      {content}
      {askRegistration && (
        <RegistrationSheet
          home={data ? data.location : null}
          onSaved={() => {
            setAskRegistration(false)
            loadClinics(false)
          }}
          onClose={() => setAskRegistration(false)}
        />
      )}
    </div>
  )
}

export default ClinicsPage
