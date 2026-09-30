import { useEffect, useState } from 'react'
import Icon from './Icon'

function CatalogSuggestions({ query, fetcher, icon, subtitle, onPick }) {
  const [items, setItems] = useState([])

  useEffect(() => {
    const q = query.trim()
    if (q.length < 2) {
      setItems([])
      return
    }

    let cancelled = false
    const timer = setTimeout(async () => {
      try {
        const data = await fetcher(q)
        if (!cancelled) {
          setItems(data || [])
        }
      } catch {
        if (!cancelled) {
          setItems([])
        }
      }
    }, 300)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [query, fetcher])

  if (items.length === 0) {
    return null
  }

  return (
    <div className="suggestions">
      {items[0].corrected && <p className="suggestions-hint">Возможно, вы имели в виду</p>}
      <ul className="suggestions-list">
        {items.map((item) => (
          <li key={(item.kind || '') + item.id}>
            <button type="button" className="suggestion" onClick={() => onPick(item)}>
              <span className={'suggestion-icon suggestion-icon-' + (item.icon || icon)}>
                <Icon name={item.icon || icon} size={20} />
              </span>
              <span className="suggestion-text">
                <span className="suggestion-name">
                  {item.name}
                  {item.matched !== item.name && <span className="suggestion-alias"> · {item.matched}</span>}
                </span>
                <span className="suggestion-sub">{item.subtitle || subtitle}</span>
              </span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default CatalogSuggestions
