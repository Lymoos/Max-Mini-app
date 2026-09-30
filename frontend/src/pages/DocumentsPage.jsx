import { useEffect, useState } from 'react'
import { getBenefits } from '../api'
import { categoryIcons, statusNames } from '../benefitStatus'
import Icon from '../components/Icon'

function DocumentsPage({ onOpenBenefit }) {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('all')

  async function loadBenefits() {
    setError('')
    try {
      setData(await getBenefits())
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadBenefits()
  }, [])

  let content
  if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить льготы</p>
        <button type="button" className="link-btn" onClick={loadBenefits}>
          Повторить
        </button>
      </div>
    )
  } else if (data === null) {
    content = <p className="card card-note">Загрузка…</p>
  } else {
    const shown = data.items.filter((b) => filter === 'all' || (filter === 'fit' && b.fit) || b.category === filter)
    const ordered = data.categories.flatMap((c) => shown.filter((b) => b.category === c))
    content = (
      <>
        <div className="doc-summary">
          <div className="doc-stat doc-stat-fit">
            <span className="doc-stat-number">{data.summary.fit}</span>
            <span className="doc-stat-label">подходит вам</span>
          </div>
          <div className="doc-stat doc-stat-progress">
            <span className="doc-stat-number">{data.summary.inProgress}</span>
            <span className="doc-stat-label">в процессе</span>
          </div>
          <div className="doc-stat doc-stat-done">
            <span className="doc-stat-number">{data.summary.approved}</span>
            <span className="doc-stat-label">одобрено</span>
          </div>
        </div>

        <div className="chips doc-chips" role="group" aria-label="Показать">
          {['all', 'fit', ...data.categories].map((f) => (
            <button
              key={f}
              type="button"
              className={filter === f ? 'chip chip-active' : 'chip'}
              onClick={() => setFilter(f)}
              aria-pressed={filter === f}
            >
              {f === 'all' ? 'Все' : f === 'fit' ? 'Подходит мне' : f}
            </button>
          ))}
        </div>

        {shown.length === 0 && (
          <p className="card card-note">
            {filter === 'fit' ? 'Укажите дату рождения в профиле — подскажем, что вам положено' : 'Здесь пока пусто'}
          </p>
        )}

        {data.categories.map((category) => {
          const list = shown.filter((b) => b.category === category)
          if (list.length === 0) {
            return null
          }
          return (
            <section key={category} className="doc-group">
              <h2 className="block-title doc-group-title">{category}</h2>
              <ul className="doc-list">
                {list.map((b) => {
                  return (
                    <li key={b.id} className="appear" style={{ animationDelay: ordered.indexOf(b) * 40 + 'ms' }}>
                      <button type="button" className="card doc-card" onClick={() => onOpenBenefit(b.id)}>
                        <span className={'doc-icon doc-icon-' + categoryIcons[b.category]}>
                          <Icon name={categoryIcons[b.category]} />
                        </span>
                        <span className="doc-main">
                          <span className="doc-title">{b.title}</span>
                          <span className="doc-short">{b.short}</span>
                          <span className="doc-badges">
                            {b.fit && <span className="badge badge-fit">{b.fitReason || 'Вам подходит'}</span>}
                            {b.auto && <span className="badge badge-auto">Без заявления</span>}
                            {b.status !== 'not_started' && (
                              <span className={'badge badge-status badge-' + b.status}>{statusNames[b.status]}</span>
                            )}
                            {b.stale && <span className="badge badge-stale">Проверьте статус</span>}
                          </span>
                          {b.status === 'collecting' && b.docsTotal > 0 && (
                            <span className="doc-progress">
                              <span className="doc-progress-bar">
                                <span className="doc-progress-fill" style={{ width: (b.docsDone / b.docsTotal) * 100 + '%' }} />
                              </span>
                              Документы: {b.docsDone} из {b.docsTotal}
                            </span>
                          )}
                        </span>
                      </button>
                    </li>
                  )
                })}
              </ul>
            </section>
          )
        })}

        <p className="disclaimer">
          Сведения справочные. Условия в регионах отличаются — уточняйте в Социальном фонде, соцзащите или на Госуслугах.
        </p>
      </>
    )
  }

  return (
    <div className="page">
      <h1 className="page-title">Документы и льготы</h1>
      {content}
    </div>
  )
}

export default DocumentsPage
