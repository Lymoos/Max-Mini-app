import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import BirthdayPicker from './BirthdayPicker'

const today = new Date(2026, 8, 30)

function renderPicker(value = '') {
  const onDone = vi.fn()
  const onClose = vi.fn()
  render(<BirthdayPicker value={value} today={today} onDone={onDone} onClose={onClose} />)
  return { onDone, onClose }
}

function cell(name) {
  return screen.getByRole('button', { name: name })
}

describe('BirthdayPicker', () => {
  it('день → месяц → год', async () => {
    const user = userEvent.setup()
    const { onDone } = renderPicker()

    expect(screen.getByText('Выберите день')).toBeInTheDocument()
    await user.click(cell('12'))
    expect(screen.getByText('Выберите месяц')).toBeInTheDocument()
    await user.click(cell('Март'))
    expect(screen.getByText('Выберите год')).toBeInTheDocument()
    expect(screen.getByText('1950 – 1959')).toBeInTheDocument()
    await user.click(cell('1956'))

    expect(onDone).toHaveBeenCalledWith('1956-03-12')
  })

  it('после 31-го нельзя выбрать месяц, где 30 дней, и февраль', async () => {
    const user = userEvent.setup()
    renderPicker()
    await user.click(cell('31'))

    for (const m of ['Февраль', 'Апрель', 'Июнь', 'Сентябрь', 'Ноябрь']) {
      expect(cell(m)).toBeDisabled()
    }
    for (const m of ['Январь', 'Март', 'Май', 'Июль', 'Август', 'Октябрь', 'Декабрь']) {
      expect(cell(m)).toBeEnabled()
    }
  })

  it('29 февраля доступно только в високосные годы', async () => {
    const user = userEvent.setup()
    renderPicker()
    await user.click(cell('29'))
    await user.click(cell('Февраль'))
    await user.click(screen.getByLabelText('Позже'))

    expect(cell('1960')).toBeEnabled()
    expect(cell('1964')).toBeEnabled()
    expect(cell('1961')).toBeDisabled()
    expect(cell('1962')).toBeDisabled()
  })

  it('будущие даты и годы до 1900 недоступны', async () => {
    const user = userEvent.setup()
    renderPicker()
    await user.click(cell('5'))
    await user.click(cell('Октябрь'))

    for (let i = 0; i < 10; i++) {
      await user.click(screen.getByLabelText('Позже'))
    }
    expect(screen.getByText('2020 – 2029')).toBeInTheDocument()
    expect(screen.getByLabelText('Позже')).toBeDisabled()
    expect(cell('2025')).toBeEnabled()
    expect(cell('2026')).toBeDisabled()
    expect(cell('2027')).toBeDisabled()

    for (let i = 0; i < 20; i++) {
      await user.click(screen.getByLabelText('Раньше'))
    }
    expect(screen.getByText('1900 – 1909')).toBeInTheDocument()
    expect(screen.getByLabelText('Раньше')).toBeDisabled()
    expect(cell('1900')).toBeEnabled()
  })

  it('открывается на сохранённой дате и можно вернуться к шагу', async () => {
    const user = userEvent.setup()
    const { onDone } = renderPicker('1981-10-05')

    expect(document.querySelector('.bday-days .bday-cell-active')).toHaveTextContent('5')
    expect(screen.getByRole('button', { name: 'октября' })).toBeInTheDocument()

    await user.click(cell('7'))
    await user.click(cell('Октябрь'))
    expect(screen.getByText('1980 – 1989')).toBeInTheDocument()
    expect(cell('1981')).toHaveClass('bday-cell-active')

    await user.click(screen.getByRole('button', { name: '7' }))
    expect(screen.getByText('Выберите день')).toBeInTheDocument()
    await user.click(cell('8'))
    await user.click(cell('Октябрь'))
    await user.click(cell('1981'))
    expect(onDone).toHaveBeenCalledWith('1981-10-08')
  })

  it('если выбрали 31-е после апреля, месяц сбрасывается', async () => {
    const user = userEvent.setup()
    renderPicker('1956-04-10')
    await user.click(cell('31'))
    expect(screen.getByRole('button', { name: 'Месяц' })).toBeInTheDocument()
    expect(cell('Апрель')).toBeDisabled()
  })

  it('«Отмена» и затемнение закрывают, клик внутри — нет', async () => {
    const user = userEvent.setup()
    const { onClose } = renderPicker()
    await user.click(screen.getByText('Выберите день'))
    expect(onClose).not.toHaveBeenCalled()
    await user.click(cell('Отмена'))
    expect(onClose).toHaveBeenCalledTimes(1)
    await user.click(document.querySelector('.overlay'))
    expect(onClose).toHaveBeenCalledTimes(2)
  })
})
