import { useEffect, useState } from 'react'
import { createTask, readRecipe } from '../api'
import { suggestVisit } from '../format'
import Icon from './Icon'

function RecipeSheet({ file, onAdded, onClose, onRetry }) {
  const [status, setStatus] = useState('reading')
  const [error, setError] = useState('')
  const [items, setItems] = useState([])
  const [visit] = useState(() => suggestVisit(new Date()))
  const [time, setTime] = useState(visit.time)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let active = true
    readRecipe(file)
      .then((data) => {
        if (active) {
          setItems((data.items || []).map((item) => ({ ...item, checked: true })))
          setStatus('done')
        }
      })
      .catch((err) => {
        if (active) {
          setError(err.message)
          setStatus('error')
        }
      })
    return () => {
      active = false
    }
  }, [file])

  function toggle(index) {
    setItems((current) => current.map((item, i) => (i === index ? { ...item, checked: !item.checked } : item)))
  }

  const chosen = items.filter((item) => item.checked)

  async function addList() {
    setSaving(true)
    setError('')
    try {
      const task = await createTask({
        title: 'Купить в аптеке',
        time: time,
        date: visit.date,
        kind: 'medicine',
        note: 'Лекарства по рецепту. Возьмите рецепт с собой — без него часть лекарств не продадут.',
        items: chosen.map((item) => ({
          title: item.dose ? item.title + ', ' + item.dose : item.title,
          medicineId: item.medicineId,
        })),
      })
      onAdded(task, visit.tomorrow)
    } catch (err) {
      setError(err.message)
      setSaving(false)
    }
  }

  let content
  if (status === 'reading') {
    content = (
      <div className="recipe-reading" role="status">
        <span className="recipe-scan">
          <Icon name="document" size={56} />
        </span>
        <p className="recipe-reading-title">Читаю рецепт…</p>
        <p className="recipe-hint">Обычно это занимает несколько секунд</p>
      </div>
    )
  } else if (status === 'error' || items.length === 0) {
    content = (
      <>
        <p className="recipe-empty">
          {status === 'error'
            ? error
            : 'Не нашёл лекарств на фото. Сфотографируйте рецепт поближе, при хорошем свете, чтобы строчки были ровными.'}
        </p>
        <div className="sheet-buttons">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Закрыть
          </button>
          <button type="button" className="btn btn-primary" onClick={onRetry}>
            Другое фото
          </button>
        </div>
      </>
    )
  } else {
    content = (
      <>
        <p className="recipe-hint">Снимите галочку, если что-то уже есть дома</p>
        <ul className="recipe-list">
          {items.map((item, i) => (
            <li key={i}>
              <label className={item.checked ? 'recipe-item recipe-item-on' : 'recipe-item'}>
                <input type="checkbox" checked={item.checked} onChange={() => toggle(i)} />
                <span className="task-circle">{item.checked && <Icon name="check" size={16} />}</span>
                <span className="recipe-item-text">
                  <span className="recipe-item-name">{item.title}</span>
                  {item.dose && <span className="recipe-item-dose">{item.dose}</span>}
                </span>
              </label>
            </li>
          ))}
        </ul>

        <label className="field recipe-time">
          <span className="field-label">{visit.tomorrow ? 'Когда пойти в аптеку завтра' : 'Когда пойти в аптеку сегодня'}</span>
          <input className="field-input" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
        </label>

        {error !== '' && <p className="error-text">{error}</p>}

        <div className="sheet-buttons recipe-buttons">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Отмена
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={addList}
            disabled={saving || chosen.length === 0 || time === ''}
          >
            {saving ? 'Добавляем…' : 'Добавить список в задачи'}
          </button>
        </div>
      </>
    )
  }

  return (
    <div className="overlay" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Рецепт по фото">
        <h2 className="sheet-title">{status === 'done' && items.length > 0 ? 'Нашёл в рецепте' : 'Рецепт по фото'}</h2>
        {content}
      </div>
    </div>
  )
}

export default RecipeSheet
