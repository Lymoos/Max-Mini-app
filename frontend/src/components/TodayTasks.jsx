import { useEffect, useRef, useState } from 'react'
import { getTodayTasks, setTaskDone, setTaskItemDone } from '../api'
import AddTaskForm from './AddTaskForm'
import Icon from './Icon'

const HIDE_DELAY = 5000

const kindIcons = {
  medicine: 'pill',
  doctor: 'stethoscope',
  call: 'phone',
}

// пока задачи грузятся, показываем серые силуэты тех же размеров — страница не прыгает
export function TasksSkeleton() {
  return (
    <ul className="task-list" aria-label="Загрузка задач">
      {[1, 2, 3].map((n) => (
        <li key={n} className="task skeleton-row">
          <span className="skeleton skeleton-icon" />
          <span className="task-text">
            <span className={'skeleton skeleton-line skeleton-line-' + n} />
            <span className="skeleton skeleton-line skeleton-line-short" />
          </span>
          <span className="skeleton skeleton-circle" />
        </li>
      ))}
    </ul>
  )
}

function progressText(task) {
  const done = task.items.filter((item) => item.done).length
  return 'отмечено ' + done + ' из ' + task.items.length
}

function TodayTasks({ onOpenMedicine }) {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [saveError, setSaveError] = useState('')
  const [busy, setBusy] = useState([])
  const [recent, setRecent] = useState([])
  const [showForm, setShowForm] = useState(false)
  const [openId, setOpenId] = useState(null)
  const timers = useRef({})

  async function loadTasks() {
    setLoading(true)
    setLoadError('')
    try {
      const data = await getTodayTasks()
      setTasks((data.tasks || []).map((t) => ({ ...t, items: t.items || [], note: t.note || '' })))
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
    const full = { ...task, items: task.items || [], note: task.note || '' }
    setTasks((current) => [...current, full].sort((a, b) => a.time.localeCompare(b.time)))
  }

  function updateItem(taskId, itemId, done) {
    setTasks((current) =>
      current.map((t) =>
        t.id === taskId ? { ...t, items: t.items.map((item) => (item.id === itemId ? { ...item, done: done } : item)) } : t,
      ),
    )
  }

  async function toggleItem(task, item) {
    setSaveError('')
    updateItem(task.id, item.id, !item.done)
    try {
      await setTaskItemDone(task.id, item.id, !item.done)
    } catch {
      updateItem(task.id, item.id, item.done)
      setSaveError('Не удалось сохранить. Попробуйте ещё раз')
    }
  }

  const visible = tasks.filter((t) => !t.done || recent.includes(t.id))

  let content
  if (loading) {
    content = <TasksSkeleton />
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
        {visible.map((task) => {
          const hasDetails = task.note !== '' || task.items.length > 0
          const open = openId === task.id
          const text = (
            <>
              <span className="task-title">{task.title}</span>
              <span className="task-time">
                {task.time}
                {task.items.length > 0 && ' · ' + progressText(task)}
              </span>
            </>
          )
          return (
            <li key={task.id} className={task.done ? 'task task-done' : 'task'}>
              <span className={'task-icon task-icon-' + task.kind}>
                <Icon name={kindIcons[task.kind] || 'calendar'} size={20} />
              </span>
              {hasDetails ? (
                <button
                  type="button"
                  className="task-text task-open"
                  onClick={() => setOpenId(open ? null : task.id)}
                  aria-expanded={open}
                >
                  {text}
                  <span className="task-more">
                    {open ? 'Свернуть' : 'Подробнее'}
                    <span className={open ? 'task-arrow task-arrow-open' : 'task-arrow'}>
                      <Icon name="down" size={18} />
                    </span>
                  </span>
                </button>
              ) : (
                <span className="task-text">{text}</span>
              )}
              <button
                type="button"
                className="task-check"
                onClick={() => toggleTask(task)}
                aria-label={(task.done ? 'Отменить отметку: ' : 'Отметить выполненной: ') + task.title}
              >
                <span className="task-circle">{task.done && <Icon name="check" size={16} />}</span>
              </button>
              {open && (
                <div className="task-details">
                  {task.note !== '' && <p className="task-note">{task.note}</p>}
                  {task.items.length > 0 && (
                    <ul className="subtask-list">
                      {task.items.map((item) => (
                        <li key={item.id} className={item.done ? 'subtask subtask-done' : 'subtask'}>
                          <button
                            type="button"
                            className="subtask-check"
                            onClick={() => toggleItem(task, item)}
                            aria-pressed={item.done}
                          >
                            <span className="task-circle">{item.done && <Icon name="check" size={16} />}</span>
                            <span className="subtask-title">{item.title}</span>
                          </button>
                          {item.medicineId && onOpenMedicine && (
                            <button
                              type="button"
                              className="link-btn subtask-prices"
                              onClick={() => onOpenMedicine({ id: item.medicineId, name: '', form: '' })}
                            >
                              Где купить
                            </button>
                          )}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              )}
            </li>
          )
        })}
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
