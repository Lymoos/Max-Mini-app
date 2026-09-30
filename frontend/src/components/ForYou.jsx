import { useEffect, useState } from 'react'
import { getForYou } from '../api'
import { formatDistance } from '../format'
import Icon from './Icon'

const STORAGE_KEY = 'forYouOpen'

function readOpen() {
  try {
    return localStorage.getItem(STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

function saveOpen(value) {
  try {
    localStorage.setItem(STORAGE_KEY, value ? '1' : '0')
  } catch {
    // в приватном режиме хранилище бывает недоступно — тогда просто не запоминаем
  }
}

function ForYou({ onOpenShop, onOpenBenefit }) {
  const [items, setItems] = useState([])
  const [open, setOpen] = useState(readOpen)

  useEffect(() => {
    getForYou()
      .then((data) => setItems(data || []))
      .catch(() => setItems([]))
  }, [])

  function toggle() {
    saveOpen(!open)
    setOpen(!open)
  }

  if (items.length === 0) {
    return null
  }

  return (
    <section className="block">
      <button type="button" className="foryou-toggle" onClick={toggle} aria-expanded={open}>
        <span className="block-title">Для вас</span>
        <span className="foryou-count">{items.length}</span>
        <span className={open ? 'foryou-arrow foryou-arrow-open' : 'foryou-arrow'}>
          <Icon name="back" size={20} />
        </span>
      </button>
      <div className={open ? 'foryou-body foryou-body-open' : 'foryou-body'}>
        <div className="foryou-inner">
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
                {item.type === 'passport' && onOpenBenefit && (
                  <button
                    type="button"
                    className="link-btn foryou-link"
                    onClick={() => onOpenBenefit('passport')}
                    tabIndex={open ? 0 : -1}
                  >
                    Как заменить паспорт
                  </button>
                )}
                {item.places && item.places.length > 0 && (
                  <ul className="foryou-places">
                    {item.places.map((p) => (
                      <li key={p.id}>
                        {p.kind === 'shop' ? (
                          <button type="button" className="foryou-place" onClick={() => onOpenShop(p)} tabIndex={open ? 0 : -1}>
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
        </div>
      </div>
    </section>
  )
}

export default ForYou
