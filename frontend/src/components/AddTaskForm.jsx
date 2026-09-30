import { useState } from 'react'
import { createTask } from '../api'

const kinds = [
  { id: 'medicine', title: 'Лекарство' },
  { id: 'doctor', title: 'Врач' },
  { id: 'call', title: 'Звонок' },
  { id: 'other', title: 'Другое' },
]

function AddTaskForm({ onAdded, onClose }) {
  const [title, setTitle] = useState('')
  const [time, setTime] = useState('')
  const [kind, setKind] = useState('other')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const canSave = title.trim() !== '' && time !== '' && !saving

  async function handleSubmit(e) {
    e.preventDefault()
    if (!canSave) {
      return
    }

    setSaving(true)
    setError('')
    try {
      const task = await createTask({ title: title, time: time, kind: kind })
      onAdded(task)
      onClose()
    } catch (err) {
      setError(err.message)
      setSaving(false)
    }
  }

  return (
    <div className="overlay" onClick={onClose}>
      <form
        className="sheet"
        onSubmit={handleSubmit}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-label="Новая задача"
      >
        <h2 className="sheet-title">Новая задача</h2>

        <label className="field">
          <span className="field-label">Что нужно сделать</span>
          <input
            className="field-input"
            type="text"
            maxLength={200}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
          />
        </label>

        <label className="field">
          <span className="field-label">Время</span>
          <input className="field-input" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
        </label>

        <div className="field">
          <span className="field-label">Вид</span>
          <div className="kinds">
            {kinds.map((k) => (
              <button
                key={k.id}
                type="button"
                className={k.id === kind ? 'kind kind-active' : 'kind'}
                onClick={() => setKind(k.id)}
                aria-pressed={k.id === kind}
              >
                {k.title}
              </button>
            ))}
          </div>
        </div>

        {error !== '' && <p className="error-text">{error}</p>}

        <div className="sheet-buttons">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" className="btn btn-primary" disabled={!canSave}>
            {saving ? 'Сохраняем…' : 'Добавить'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default AddTaskForm
