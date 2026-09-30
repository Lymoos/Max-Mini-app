import { useEffect, useRef, useState } from 'react'
import { getBenefit, saveBenefit } from '../api'
import { categoryIcons, statusNames, stepIndex } from '../benefitStatus'
import Icon from '../components/Icon'
import { formatFullDate, phoneLink } from '../format'

const stepTitles = ['Собираю', 'Подано', 'Рассматривают', 'Решение']

function BenefitPage({ benefitId, onBack }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [saveError, setSaveError] = useState('')
  const [saving, setSaving] = useState(false)
  const docsRef = useRef([])
  const docsQueue = useRef(Promise.resolve())

  async function loadBenefit() {
    setError('')
    try {
      const result = await getBenefit(benefitId)
      docsRef.current = result.state.docs
      setData(result)
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadBenefit()
  }, [benefitId])

  async function change(changes, preview) {
    if (saving) {
      return
    }
    const old = data.state
    setSaving(true)
    setSaveError('')
    setData({ ...data, state: { ...old, ...preview } })
    try {
      const state = await saveBenefit(benefitId, changes)
      setData((current) => ({ ...current, state: { ...state, docs: docsRef.current }, stale: false }))
    } catch (err) {
      setData((current) => ({ ...current, state: { ...old, docs: docsRef.current } }))
      setSaveError(err.message)
    }
    setSaving(false)
  }

  // галочки меняются сразу, а запросы уходят по очереди, чтобы быстрые нажатия не терялись
  function toggleDoc(i) {
    const current = docsRef.current
    const docs = current.includes(i) ? current.filter((d) => d !== i) : [...current, i]
    docsRef.current = docs
    setSaveError('')
    setData((d) => ({ ...d, state: { ...d.state, docs: docs } }))
    docsQueue.current = docsQueue.current
      .then(() => saveBenefit(benefitId, { docs: docs }))
      .catch((err) => {
        setSaveError(err.message)
        loadBenefit()
      })
  }

  function setStatus(status) {
    change({ status: status }, { status: status })
  }

  if (error !== '') {
    return (
      <div className="page">
        <button type="button" className="back-btn" onClick={onBack}>
          <Icon name="back" size={20} />
          Назад
        </button>
        <div className="card card-note">
          <p>Не удалось загрузить</p>
          <button type="button" className="link-btn" onClick={loadBenefit}>
            Повторить
          </button>
        </div>
      </div>
    )
  }
  if (data === null) {
    return (
      <div className="page">
        <p className="card card-note">Загрузка…</p>
      </div>
    )
  }

  const b = data.benefit
  const st = data.state
  const step = stepIndex(st.status)
  const docsDone = st.docs.length

  let actions
  if (st.status === 'not_started') {
    actions = (
      <button type="button" className="btn btn-primary status-btn" onClick={() => setStatus('collecting')} disabled={saving}>
        Начать: собираю документы
      </button>
    )
  } else if (st.status === 'collecting') {
    actions = (
      <button type="button" className="btn btn-primary status-btn" onClick={() => setStatus('submitted')} disabled={saving}>
        Я подал(а) заявление
      </button>
    )
  } else if (st.status === 'submitted') {
    actions = (
      <button type="button" className="btn btn-primary status-btn" onClick={() => setStatus('review')} disabled={saving}>
        Заявление приняли на рассмотрение
      </button>
    )
  } else if (st.status === 'review') {
    actions = (
      <div className="status-two">
        <button type="button" className="btn btn-primary status-btn status-ok" onClick={() => setStatus('approved')} disabled={saving}>
          Одобрили
        </button>
        <button type="button" className="btn btn-secondary status-btn" onClick={() => setStatus('rejected')} disabled={saving}>
          Отказали
        </button>
      </div>
    )
  } else {
    actions = (
      <button type="button" className="link-btn status-reset" onClick={() => setStatus('not_started')} disabled={saving}>
        Начать заново
      </button>
    )
  }

  return (
    <div className="page benefit-page">
      <button type="button" className="back-btn" onClick={onBack}>
        <Icon name="back" size={20} />
        Назад
      </button>

      <div className="benefit-head appear">
        <span className={'doc-icon doc-icon-big doc-icon-' + categoryIcons[b.category]}>
          <Icon name={categoryIcons[b.category]} size={28} />
        </span>
        <h1 className="page-title benefit-title">{b.title}</h1>
        <p className="benefit-short">{b.short}</p>
        {data.fit && <span className="badge badge-fit">{data.fitReason}</span>}
      </div>

      <div className="card benefit-block appear">
        <h2 className="benefit-block-title">Кому положено</h2>
        <p className="benefit-text">{b.who}</p>
      </div>

      {b.deadline && (
        <div className="card benefit-deadline appear">
          <Icon name="calendar" size={20} />
          <p>{b.deadline}</p>
        </div>
      )}

      {b.auto ? (
        <div className="card benefit-auto appear">
          <Icon name="check" size={22} />
          <p>Назначается автоматически — подавать заявление не нужно</p>
        </div>
      ) : (
        <div className="card benefit-block appear">
          <h2 className="benefit-block-title">Статус заявления</h2>
          <div className="stepper" aria-label={'Статус: ' + statusNames[st.status]}>
            <div className="stepper-line">
              <div
                className={st.status === 'rejected' ? 'stepper-fill stepper-fill-bad' : 'stepper-fill'}
                style={{ width: step <= 0 ? '0%' : (step / 3) * 100 + '%' }}
              />
            </div>
            {stepTitles.map((title, i) => (
              <div key={title} className={i <= step ? 'stepper-step stepper-step-done' : 'stepper-step'}>
                <span className="stepper-dot">{i < step || st.status === 'approved' ? <Icon name="check" size={14} /> : i + 1}</span>
                <span className="stepper-title">{i === 3 && st.status === 'rejected' ? 'Отказ' : title}</span>
              </div>
            ))}
          </div>
          <p className={'status-now status-now-' + st.status}>{statusNames[st.status]}</p>
          {st.submittedAt && <p className="status-date">Подано {formatFullDate(st.submittedAt)}</p>}
          {data.stale && (
            <p className="status-stale">Прошло больше месяца — проверьте статус заявления на Госуслугах</p>
          )}
          {st.status === 'rejected' && (
            <p className="status-hint">Причина отказа указана в решении. Её можно исправить и подать заявление снова.</p>
          )}
          {actions}
          {saveError !== '' && <p className="error-text">{saveError}</p>}
        </div>
      )}

      {b.documents && b.documents.length > 0 && (
        <div className="card benefit-block appear">
          <h2 className="benefit-block-title">
            Документы <span className="docs-count">{docsDone} из {b.documents.length}</span>
          </h2>
          <ul className="checklist">
            {b.documents.map((doc, i) => {
              const checked = st.docs.includes(i)
              return (
                <li key={doc}>
                  <button
                    type="button"
                    className={checked ? 'check-item check-item-done' : 'check-item'}
                    onClick={() => toggleDoc(i)}
                    aria-pressed={checked}
                  >
                    <span className="check-box">{checked && <Icon name="check" size={16} />}</span>
                    <span className="check-text">{doc}</span>
                  </button>
                </li>
              )
            })}
          </ul>
        </div>
      )}

      <div className="card benefit-block appear">
        <h2 className="benefit-block-title">Как {b.auto ? 'проверить' : 'подать'}</h2>
        <ol className="how-steps">
          {b.steps.map((stepText, i) => (
            <li key={i} className="how-step">
              <span className="how-num">{i + 1}</span>
              <span className="how-text">{stepText}</span>
            </li>
          ))}
        </ol>
      </div>

      <div className="benefit-links appear">
        {b.links.map((link, i) => (
          <a
            key={link.url}
            className={i === 0 ? 'btn btn-primary benefit-link' : 'btn btn-secondary benefit-link'}
            href={link.url}
            target="_blank"
            rel="noreferrer"
          >
            {link.title}
            <Icon name="external" size={18} />
          </a>
        ))}
        {b.phone && (
          <a className="btn btn-secondary benefit-link" href={phoneLink(b.phone)}>
            <Icon name="phone" size={18} />
            Социальный фонд: {b.phone}
          </a>
        )}
      </div>

      <p className="disclaimer">Сведения справочные. Условия уточняйте в Социальном фонде, соцзащите или на Госуслугах.</p>
    </div>
  )
}

export default BenefitPage
