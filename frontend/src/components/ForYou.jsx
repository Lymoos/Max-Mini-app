import { useEffect, useState } from 'react'
import { getForYou } from '../api'
import { formatDistance } from '../format'
import Icon from './Icon'

function ForYou({ onOpenShop }) {
  const [items, setItems] = useState([])

  useEffect(() => {
    getForYou()
      .then((data) => setItems(data || []))
      .catch(() => setItems([]))
  }, [])

  if (items.length === 0) {
    return null
  }

  return (
    <section className="block">
      <div className="block-header">
        <h2 className="block-title">Для вас</h2>
      </div>
      <div className="foryou-list">
        {items.map((item) => (
          <div key={item.type} className={'card foryou foryou-' + item.type}>
            <div className="foryou-head">
              <span className="foryou-icon">
                <Icon name={item.type === 'birthday' ? 'gift' : 'document'} />
              </span>
              <p className="foryou-title">{item.title}</p>
            </div>
            <p className="foryou-text">{item.text}</p>
            {item.places && item.places.length > 0 && (
              <ul className="foryou-places">
                {item.places.map((p) => (
                  <li key={p.id}>
                    {p.kind === 'shop' ? (
                      <button type="button" className="foryou-place" onClick={() => onOpenShop(p)}>
                        <span className="foryou-place-name">{p.name}</span>
                        <span className="foryou-place-info">
                          −{p.birthdayDiscount}% · {formatDistance(p.distanceKm)}
                        </span>
                      </button>
                    ) : (
                      <div className="foryou-place">
                        <span className="foryou-place-name">{p.name}</span>
                        <span className="foryou-place-info">
                          −{p.birthdayDiscount}% · {formatDistance(p.distanceKm)}
                        </span>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

export default ForYou
