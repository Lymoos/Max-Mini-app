import { useEffect, useRef, useState } from 'react'
import { getTodayTasks, setTaskDone } from '../api'
import AddTaskForm from './AddTaskForm'
import Icon from './Icon'

const HIDE_DELAY = 5000

const kindIcons = {
  medicine: 'pill',
  doctor: 'stethoscope',
  call: 'phone',
}

function TodayTasks() {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [saveError, setSaveError] = useState('')
  const [busy, setBusy] = useState([])
  const [recent, setRecent] = useState([])
  const [showForm, setShowForm] = useState(false)
  const timers = useRef({})

  async function loadTasks() {
    setLoading(true)
    setLoadError('')
    try {
      const data = await getTodayTasks()
      setTasks(data.tasks || [])
    } catch (err) {
      setLoadError(err.message)
    }
    setLoading(false)
  }

  useEffect(() => {
    loadTasks()
    const allTimers = timers.current
    return () => {
      Object.values(allTimers).forEach(clearTimeout)
    }
  }, [])

  function updateDone(id, done) {
    setTasks((current) => current.map((t) => (t.id === id ? { ...t, done: done } : t)))
  }

  function startHide(id) {
    clearTimeout(timers.current[id])
    setRecent((current) => (current.includes(id) ? current : [...current, id]))
    timers.current[id] = setTimeout(() => {
      delete timers.current[id]
      setRecent((current) => current.filter((x) => x !== id))
    }, HIDE_DELAY)
  }

  function stopHide(id) {
    clearTimeout(timers.current[id])
    delete timers.current[id]
    setRecent((current) => current.filter((x) => x !== id))
  }

  async function toggleTask(task) {
    if (busy.includes(task.id)) {
      return
    }
    const newDone = !task.done
    setBusy((current) => [...current, task.id])
    setSaveError('')
    updateDone(task.id, newDone)
    if (newDone) {
      startHide(task.id)
    } else {
      stopHide(task.id)
    }

    try {
      await setTaskDone(task.id, newDone)
    } catch {
      updateDone(task.id, task.done)
      if (task.done) {
        startHide(task.id)
      } else {
        stopHide(task.id)
      }
      setSaveError('Не удалось сохранить. Попробуйте ещё раз')
    }
    setBusy((current) => current.filter((x) => x !== task.id))
  }

  function handleAdded(task) {
    setTasks((current) => [...current, task].sort((a, b) => a.time.localeCompare(b.time)))
  }

  const visible = tasks.filter((t) => !t.done || recent.includes(t.id))

  let content
  if (loading) {
    content = <p className="tasks-note">Загрузка…</p>
  } else if (loadError !== '') {
    content = (
      <div className="tasks-note">
        <p>Не удалось загрузить задачи</p>
        <button type="button" className="link-btn" onClick={loadTasks}>
          Повторить
        </button>
      </div>
    )
  } else if (visible.length === 0) {
    content = (
      <p className="tasks-note">{tasks.length > 0 ? 'Все задачи на сегодня выполнены' : 'На сегодня задач нет'}</p>
    )
  } else {
    content = (
      <ul className="task-list">
        {visible.map((task) => (
          <li key={task.id} className={task.done ? 'task task-done' : 'task'}>
            <span className={'task-icon task-icon-' + task.kind}>
              <Icon name={kindIcons[task.kind] || 'calendar'} size={20} />
            </span>
            <span className="task-text">
              <span className="task-title">{task.title}</span>
              <span className="task-time">{task.time}</span>
            </span>
            <button
              type="button"
              className="task-check"
              onClick={() => toggleTask(task)}
              aria-label={(task.done ? 'Отменить отметку: ' : 'Отметить выполненной: ') + task.title}
            >
              <span className="task-circle">{task.done && <Icon name="check" size={16} />}</span>
            </button>
          </li>
        ))}
      </ul>
    )
  }

  return (
    <section className="block tasks-block">
      <div className="tasks-header">
        <h2 className="tasks-title">Задачи на сегодня</h2>
        <button type="button" className="add-btn" onClick={() => setShowForm(true)} aria-label="Добавить задачу">
          <Icon name="plus" size={22} />
        </button>
      </div>
      {content}
      {saveError !== '' && <p className="error-text">{saveError}</p>}
      {showForm && <AddTaskForm onAdded={handleAdded} onClose={() => setShowForm(false)} />}
    </section>
  )
}

export default TodayTasks
