import { useRef, useState } from 'react'
import { createTask, sendAsk, suggestMedicines, suggestProducts } from '../api'
import CatalogSuggestions from './CatalogSuggestions'
import Icon from './Icon'
import RecipeSheet from './RecipeSheet'

// на главной ищем сразу и лекарства, и товары
async function suggestAll(query) {
  const [meds, goods] = await Promise.all([suggestMedicines(query), suggestProducts(query)])
  const result = []
  meds.forEach((m) => result.push({ ...m, kind: 'medicine', icon: 'pill', subtitle: 'Найти в ближайших аптеках' }))
  goods.forEach((p) => result.push({ ...p, kind: 'product', icon: 'cart', subtitle: 'Найти в магазинах рядом' }))
  result.sort((a, b) => Number(a.corrected) - Number(b.corrected))
  return result.slice(0, 6)
}

function AskBox({ onAnswer, onOpenProfile, onTaskAdded }) {
  const [text, setText] = useState('')
  const [message, setMessage] = useState('')
  const [answer, setAnswer] = useState(null)
  const [taskTime, setTaskTime] = useState('')
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(false)
  const [showHints, setShowHints] = useState(false)
  const [recipeFile, setRecipeFile] = useState(null)
  const fileInput = useRef(null)

  async function handleSubmit(e) {
    e.preventDefault()
    if (text.trim() === '' || loading) {
      return
    }

    setLoading(true)
    setMessage('')
    setAnswer(null)
    setShowHints(false)
    try {
      const result = await sendAsk(text)
      if (result.type === 'unknown') {
        setMessage(result.message)
      } else {
        setAnswer(result)
        setTaskTime(result.task ? result.task.time : '')
      }
    } catch (err) {
      setMessage(err.message)
    }
    setLoading(false)
  }

  // задачу создаём только после того, как человек нажал «Добавить»
  async function addTask() {
    setSaving(true)
    try {
      const task = await createTask({ ...answer.task, time: taskTime })
      setAnswer(null)
      setText('')
      setMessage('Задача «' + task.title + '» добавлена. Напомню в ' + taskTime)
      if (onTaskAdded) {
        onTaskAdded(task)
      }
    } catch (err) {
      setMessage(err.message)
    }
    setSaving(false)
  }

  // профиль открывается поверх главной, поэтому карточку и текст убираем сами
  function handleGo() {
    if (answer.type === 'task') {
      addTask()
      return
    }
    const current = answer
    setAnswer(null)
    setText('')
    onAnswer(current)
  }

  // сразу открываем выбор фото из галереи, без промежуточных экранов
  function pickPhoto() {
    setRecipeFile(null)
    fileInput.current.value = ''
    fileInput.current.click()
  }

  function handleFile(e) {
    const file = e.target.files && e.target.files[0]
    if (file) {
      setMessage('')
      setAnswer(null)
      setRecipeFile(file)
    }
  }

  function handleRecipeAdded(task, tomorrow) {
    setRecipeFile(null)
    setMessage('Список добавлен в задачи: «' + task.title + '», ' + (tomorrow ? 'завтра' : 'сегодня') + ' в ' + task.time)
    if (onTaskAdded) {
      onTaskAdded(task)
    }
  }

  function handleChange(e) {
    setText(e.target.value)
    setShowHints(true)
  }

  function handleClear() {
    setText('')
    setMessage('')
    setAnswer(null)
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
        <button type="button" className="photo-btn" onClick={pickPhoto} aria-label="Прочитать рецепт по фото">
          <Icon name="camera" size={26} />
        </button>
        <input
          ref={fileInput}
          className="file-input"
          type="file"
          accept="image/*,application/pdf"
          onChange={handleFile}
          data-testid="recipe-file"
        />
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
      {recipeFile && (
        <RecipeSheet file={recipeFile} onAdded={handleRecipeAdded} onClose={() => setRecipeFile(null)} onRetry={pickPhoto} />
      )}
      {answer && (
        <div className="answer-card" role="status">
          <p className="answer-text">{answer.message}</p>
          {answer.type === 'task' && (
            <label className="answer-time">
              Время
              <input
                className="field-input"
                type="time"
                value={taskTime}
                onChange={(e) => setTaskTime(e.target.value)}
              />
            </label>
          )}
          <div className="answer-buttons">
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleGo}
              disabled={saving || (answer.type === 'task' && taskTime === '')}
            >
              {answer.button}
            </button>
            <button type="button" className="btn btn-secondary answer-cancel" onClick={() => setAnswer(null)}>
              Отмена
            </button>
          </div>
        </div>
      )}
    </>
  )
}

export default AskBox
