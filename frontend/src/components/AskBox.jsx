import { useState } from 'react'
import { sendAsk, suggestMedicines, suggestProducts } from '../api'
import CatalogSuggestions from './CatalogSuggestions'
import Icon from './Icon'

// на главной ищем сразу и лекарства, и товары
async function suggestAll(query) {
  const [meds, goods] = await Promise.all([suggestMedicines(query), suggestProducts(query)])
  const result = []
  meds.forEach((m) => result.push({ ...m, kind: 'medicine', icon: 'pill', subtitle: 'Найти в ближайших аптеках' }))
  goods.forEach((p) => result.push({ ...p, kind: 'product', icon: 'cart', subtitle: 'Найти в магазинах рядом' }))
  result.sort((a, b) => Number(a.corrected) - Number(b.corrected))
  return result.slice(0, 6)
}

function AskBox({ onAnswer, onOpenProfile }) {
  const [text, setText] = useState('')
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [showHints, setShowHints] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    if (text.trim() === '' || loading) {
      return
    }

    setLoading(true)
    setMessage('')
    setShowHints(false)
    try {
      const answer = await sendAsk(text)
      if (answer.type === 'unknown') {
        setMessage(answer.message)
      } else {
        onAnswer(answer)
      }
    } catch (err) {
      setMessage(err.message)
    }
    setLoading(false)
  }

  function handleChange(e) {
    setText(e.target.value)
    setShowHints(true)
  }

  function handleClear() {
    setText('')
    setMessage('')
    setShowHints(false)
  }

  function handlePick(item) {
    setShowHints(false)
    if (item.kind === 'product') {
      onAnswer({ type: 'product', target: item.id, product: item })
    } else {
      onAnswer({ type: 'medicine', target: item.id, medicine: item })
    }
  }

  return (
    <>
      <div className="search">
        <form className="search-box" onSubmit={handleSubmit} role="search">
          <span className="search-icon">
            <Icon name="search" />
          </span>
          <input
            className="search-input"
            type="text"
            placeholder="Что вам нужно?"
            aria-label="Что вам нужно?"
            enterKeyHint="search"
            maxLength={500}
            value={text}
            onChange={handleChange}
          />
          {text !== '' && (
            <button type="button" className="search-clear" onClick={handleClear} aria-label="Очистить">
              <Icon name="close" />
            </button>
          )}
        </form>
        <button type="button" className="profile-btn" onClick={onOpenProfile} aria-label="Профиль">
          <Icon name="user" />
        </button>
        {showHints && <CatalogSuggestions query={text} fetcher={suggestAll} onPick={handlePick} />}
      </div>
      {message !== '' && <p className="search-message">{message}</p>}
    </>
  )
}

export default AskBox
