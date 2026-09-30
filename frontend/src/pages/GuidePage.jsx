import { useEffect, useState } from 'react'
import { getGuide, markGuide } from '../api'
import Icon from '../components/Icon'
import PlatformSwitch from '../components/PlatformSwitch'
import { detectPlatform, savePlatform } from '../platform'

function GuidePage({ guideId, onBack }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [platform, setPlatform] = useState(detectPlatform)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')

  async function loadGuide() {
    setError('')
    try {
      setData(await getGuide(guideId))
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadGuide()
  }, [guideId])

  function changePlatform(value) {
    savePlatform(value)
    setPlatform(value)
  }

  async function setRead(read) {
    setSaving(true)
    setSaveError('')
    try {
      await markGuide(guideId, read)
      if (read) {
        onBack()
        return
      }
      setData({ ...data, read: false })
    } catch (err) {
      setSaveError(err.message)
    }
    setSaving(false)
  }

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить инструкцию</p>
        <button type="button" className="link-btn" onClick={loadGuide}>
          Повторить
        </button>
      </div>
    )
  } else if (data === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const g = data.guide
    const perPlatform = !g.steps || g.steps.length === 0
    const steps = perPlatform ? (platform === 'iphone' ? g.iphone : g.android) : g.steps
    content = (
      <>
        <div className={g.important ? 'guide-head guide-head-important appear' : 'guide-head appear'}>
          <span className="guide-icon guide-icon-big">
            <Icon name={g.icon} size={28} />
          </span>
          <h1 className="page-title benefit-title">{g.title}</h1>
        </div>
        {perPlatform && <PlatformSwitch value={platform} onChange={changePlatform} />}
        <ol className="guide-steps" key={platform}>
          {steps.map((text, i) => (
            <li key={i} className="card guide-step appear" style={{ animationDelay: i * 70 + 'ms' }}>
              <span className="how-num">{i + 1}</span>
              <span className="guide-step-text">{text}</span>
            </li>
          ))}
        </ol>
        {g.tip && (
          <div className="card guide-tip appear">
            <Icon name="help" size={20} />
            <p>{g.tip}</p>
          </div>
        )}
        {saveError !== '' && <p className="error-text">{saveError}</p>}
        {data.read ? (
          <div className="guide-done">
            <p className="saved-note">Вы уже прочитали эту инструкцию</p>
            <button type="button" className="link-btn" onClick={() => setRead(false)} disabled={saving}>
              Отметить непрочитанной
            </button>
          </div>
        ) : (
          <button type="button" className="btn btn-primary guide-ok" onClick={() => setRead(true)} disabled={saving}>
            Понятно
          </button>
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

export default GuidePage
