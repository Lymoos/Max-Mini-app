import { useEffect, useState } from 'react'
import { getFeatures } from '../api'
import Icon from './Icon'

function Features({ onOpen }) {
  const [features, setFeatures] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  async function loadFeatures() {
    setLoading(true)
    setError('')
    try {
      const data = await getFeatures()
      setFeatures(data || [])
    } catch (err) {
      setError(err.message)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadFeatures()
  }, [])

  let content
  if (loading) {
    content = <p className="card card-note">Загрузка…</p>
  } else if (error !== '') {
    content = (
      <div className="card card-note">
        <p>Не удалось загрузить</p>
        <button type="button" className="link-btn" onClick={loadFeatures}>
          Повторить
        </button>
      </div>
    )
  } else {
    content = (
      <div className="tiles">
        {features.map((f) => (
          <button key={f.id} type="button" className={'tile tile-' + f.color} onClick={() => onOpen(f)}>
            <span className="tile-title">{f.title}</span>
            <span className="tile-icon">
              <Icon name={f.icon} size={84} />
            </span>
          </button>
        ))}
      </div>
    )
  }

  return (
    <section className="block">
      <div className="block-header">
        <h2 className="block-title">Возможности</h2>
      </div>
      {content}
    </section>
  )
}

export default Features
