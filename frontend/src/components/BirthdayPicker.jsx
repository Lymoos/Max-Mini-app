import { useState } from 'react'
import Icon from './Icon'

const months = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь']

const monthsGenitive = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
  'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']

const days = []
for (let d = 1; d <= 31; d++) {
  days.push(d)
}

const MIN_YEAR = 1900

function isLeap(year) {
  return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0
}

// если год ещё не выбран, февралю разрешаем 29 дней
function daysInMonth(month, year) {
  if (month === 2) {
    return year && !isLeap(year) ? 28 : 29
  }
  return [4, 6, 9, 11].includes(month) ? 30 : 31
}

function pad(n) {
  return n < 10 ? '0' + n : String(n)
}

function BirthdayPicker({ value, today, onDone, onClose }) {
  const parts = value ? value.split('-').map(Number) : null
  const [step, setStep] = useState('day')
  const [day, setDay] = useState(parts ? parts[2] : null)
  const [month, setMonth] = useState(parts ? parts[1] : null)
  const [decade, setDecade] = useState(parts ? Math.floor(parts[0] / 10) * 10 : 1950)

  const thisYear = today.getFullYear()
  const lastDecade = Math.floor(thisYear / 10) * 10

  function pickDay(d) {
    setDay(d)
    if (month && d > daysInMonth(month)) {
      setMonth(null)
    }
    setStep('month')
  }

  function pickMonth(m) {
    setMonth(m)
    setStep('year')
  }

  function yearDisabled(year) {
    if (year < MIN_YEAR || year > thisYear) {
      return true
    }
    if (day > daysInMonth(month, year)) {
      return true
    }
    return new Date(year, month - 1, day) > today
  }

  const years = []
  for (let y = decade; y < decade + 10; y++) {
    years.push(y)
  }

  let title
  let content
  if (step === 'day') {
    title = 'Выберите день'
    content = (
      <div className="bday-grid bday-days">
        {days.map((d) => (
          <button
            key={d}
            type="button"
            className={d === day ? 'bday-cell bday-cell-active' : 'bday-cell'}
            onClick={() => pickDay(d)}
          >
            {d}
          </button>
        ))}
      </div>
    )
  } else if (step === 'month') {
    title = 'Выберите месяц'
    content = (
      <div className="bday-grid bday-months">
        {months.map((name, i) => (
          <button
            key={name}
            type="button"
            className={i + 1 === month ? 'bday-cell bday-cell-active' : 'bday-cell'}
            onClick={() => pickMonth(i + 1)}
            disabled={day > daysInMonth(i + 1)}
          >
            {name}
          </button>
        ))}
      </div>
    )
  } else {
    title = 'Выберите год'
    content = (
      <>
        <div className="bday-decade">
          <button
            type="button"
            className="bday-arrow"
            onClick={() => setDecade(decade - 10)}
            disabled={decade <= MIN_YEAR}
            aria-label="Раньше"
          >
            <Icon name="back" size={22} />
          </button>
          <span className="bday-decade-text">
            {decade} – {decade + 9}
          </span>
          <button
            type="button"
            className="bday-arrow bday-arrow-next"
            onClick={() => setDecade(decade + 10)}
            disabled={decade >= lastDecade}
            aria-label="Позже"
          >
            <Icon name="back" size={22} />
          </button>
        </div>
        <div className="bday-grid bday-years">
          {years.map((y) => (
            <button
              key={y}
              type="button"
              className={parts && y === parts[0] ? 'bday-cell bday-cell-active' : 'bday-cell'}
              onClick={() => onDone(y + '-' + pad(month) + '-' + pad(day))}
              disabled={yearDisabled(y)}
            >
              {y}
            </button>
          ))}
        </div>
      </>
    )
  }

  return (
    <div className="overlay overlay-top" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Дата рождения">
        <div className="bday-steps">
          <button
            type="button"
            className={step === 'day' ? 'bday-step bday-step-active' : 'bday-step'}
            onClick={() => setStep('day')}
          >
            {day ? day : 'День'}
          </button>
          <button
            type="button"
            className={step === 'month' ? 'bday-step bday-step-active' : 'bday-step'}
            onClick={() => setStep('month')}
            disabled={!day}
          >
            {month ? monthsGenitive[month - 1] : 'Месяц'}
          </button>
          <button
            type="button"
            className={step === 'year' ? 'bday-step bday-step-active' : 'bday-step'}
            disabled
          >
            Год
          </button>
        </div>
        <h2 className="sheet-title">{title}</h2>
        {content}
        <button type="button" className="btn btn-secondary bday-cancel" onClick={onClose}>
          Отмена
        </button>
      </div>
    </div>
  )
}

export default BirthdayPicker
