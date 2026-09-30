import { useEffect, useState } from 'react'
import { getGuides } from '../api'
import Icon from '../components/Icon'

function HelpPage({ onOpenGuide }) {
  const [guides, setGuides] = useState(null)
  const [error, setError] = useState('')

  async function loadGuides() {
    setError('')
    try {
      setGuides(await getGuides())
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadGuides()
  }, [])

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить инструкции</p>
        <button type="button" className="link-btn" onClick={loadGuides}>
          Повторить
        </button>
      </div>
    )
  } else if (guides === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const readCount = guides.filter((g) => g.read).length
    content = (
      <>
        <div className="card help-progress appear">
          <p className="help-progress-text">
            Прочитано {readCount} из {guides.length}
          </p>
          <div className="doc-progress-bar help-progress-bar">
            <span className="doc-progress-fill" style={{ width: (readCount / guides.length) * 100 + '%' }} />
          </div>
        </div>
        <ul className="doc-list">
          {guides.map((g, i) => (
            <li key={g.id} className="appear" style={{ animationDelay: i * 40 + 'ms' }}>
              <button
                type="button"
                className={g.important ? 'card guide-card guide-card-important' : 'card guide-card'}
                onClick={() => onOpenGuide(g.id)}
              >
                <span className="guide-icon">
                  <Icon name={g.icon} />
                </span>
                <span className="doc-main">
                  <span className="doc-title">{g.title}</span>
                  <span className="doc-short">{g.short}</span>
                </span>
                {g.read && (
                  <span className="guide-read" aria-label="Прочитано">
                    <Icon name="check" size={16} />
                  </span>
                )}
              </button>
            </li>
          ))}
        </ul>
      </>
    )
  }

  return (
    <div className="page">
      <h1 className="page-title">Помощь</h1>
      <p className="page-hint">Простые инструкции, как пользоваться телефоном</p>
      <div className="section-gap">{content}</div>
    </div>
  )
}

export default HelpPage
