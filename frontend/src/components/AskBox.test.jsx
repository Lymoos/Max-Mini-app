import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { jsonResponse, mockFetch } from '../testUtils'
import AskBox from './AskBox'

describe('AskBox', () => {
  it('крестик появляется только когда есть текст и очищает поле', async () => {
    const user = userEvent.setup()
    render(<AskBox onAnswer={() => {}} />)

    expect(screen.queryByLabelText('Очистить')).not.toBeInTheDocument()
    await user.type(screen.getByLabelText('Что вам нужно?'), 'аптека')
    expect(screen.getByLabelText('Очистить')).toBeInTheDocument()

    await user.click(screen.getByLabelText('Очистить'))
    expect(screen.getByLabelText('Что вам нужно?')).toHaveValue('')
    expect(screen.queryByLabelText('Очистить')).not.toBeInTheDocument()
  })

  it('пустой запрос и одни пробелы не отправляются', async () => {
    const user = userEvent.setup()
    const fetchMock = mockFetch({})
    render(<AskBox onAnswer={() => {}} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), '{Enter}')
    await user.type(screen.getByLabelText('Что вам нужно?'), '   {Enter}')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('отправляет текст и передаёт найденный раздел наверх', async () => {
    const user = userEvent.setup()
    const answer = { type: 'tab', target: 'medicines', message: 'Открываю' }
    const fetchMock = mockFetch({ 'POST /api/ask': () => jsonResponse(answer) })
    const onAnswer = vi.fn()
    render(<AskBox onAnswer={onAnswer} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'таблетки{Enter}')

    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ text: 'таблетки' })
    expect(onAnswer).toHaveBeenCalledWith(answer)
  })

  it('показывает подсказку, если запрос не понят', async () => {
    const user = userEvent.setup()
    mockFetch({ 'POST /api/ask': () => jsonResponse({ type: 'unknown', message: 'Не поняли запрос' }) })
    const onAnswer = vi.fn()
    render(<AskBox onAnswer={onAnswer} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'погода{Enter}')

    expect(await screen.findByText('Не поняли запрос')).toBeInTheDocument()
    expect(onAnswer).not.toHaveBeenCalled()
  })

  it('показывает ошибку сервера и нет связи', async () => {
    const user = userEvent.setup()
    mockFetch({ 'POST /api/ask': () => jsonResponse({ error: 'пустой запрос' }, 400) })
    render(<AskBox onAnswer={() => {}} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'x{Enter}')
    expect(await screen.findByText('пустой запрос')).toBeInTheDocument()

    mockFetch({})
    await user.type(screen.getByLabelText('Что вам нужно?'), 'y{Enter}')
    expect(await screen.findByText('Нет связи с сервером')).toBeInTheDocument()
  })

  it('кнопка профиля', async () => {
    const user = userEvent.setup()
    const onOpenProfile = vi.fn()
    render(<AskBox onAnswer={() => {}} onOpenProfile={onOpenProfile} />)
    await user.click(screen.getByLabelText('Профиль'))
    expect(onOpenProfile).toHaveBeenCalled()
  })

  it('подсказки ищут и лекарства, и товары; точные совпадения выше исправленных', async () => {
    const user = userEvent.setup()
    const med = { id: 'ibuprofen', name: 'Ибупрофен', form: 'таблетки', matched: 'Нурофен', corrected: true }
    const milk = { id: 'milk', name: 'Молоко 2,5%', unit: '1 л', matched: 'Молоко', corrected: false }
    mockFetch({
      ['GET /api/medicines/suggest?q=' + encodeURIComponent('нур')]: () => jsonResponse([med]),
      ['GET /api/products/suggest?q=' + encodeURIComponent('нур')]: () => jsonResponse([milk]),
    })
    render(<AskBox onAnswer={() => {}} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'нур')
    await screen.findByText('Ибупрофен')
    const names = [...document.querySelectorAll('.suggestion-name')].map((el) => el.firstChild.textContent)
    expect(names).toEqual(['Молоко 2,5%', 'Ибупрофен'])
    expect(screen.getByText('Найти в магазинах рядом')).toBeInTheDocument()
    expect(screen.getByText('Найти в ближайших аптеках')).toBeInTheDocument()
  })

  it('выбор лекарства и товара из подсказок', async () => {
    const user = userEvent.setup()
    const med = { id: 'ibuprofen', name: 'Ибупрофен', form: 'таблетки', matched: 'Нурофен', corrected: true }
    const milk = { id: 'milk', name: 'Молоко 2,5%', unit: '1 л', matched: 'Молоко', corrected: false }
    mockFetch({
      ['GET /api/medicines/suggest?q=' + encodeURIComponent('нурафен')]: () => jsonResponse([med]),
      ['GET /api/products/suggest?q=' + encodeURIComponent('нурафен')]: () => jsonResponse([]),
      ['GET /api/medicines/suggest?q=' + encodeURIComponent('молоко')]: () => jsonResponse([]),
      ['GET /api/products/suggest?q=' + encodeURIComponent('молоко')]: () => jsonResponse([milk]),
    })
    const onAnswer = vi.fn()
    render(<AskBox onAnswer={onAnswer} />)

    await user.type(screen.getByLabelText('Что вам нужно?'), 'нурафен')
    await user.click(await screen.findByRole('button', { name: /Ибупрофен/ }))
    expect(onAnswer).toHaveBeenLastCalledWith(expect.objectContaining({ type: 'medicine', target: 'ibuprofen' }))
    expect(screen.queryByText('Найти в ближайших аптеках')).not.toBeInTheDocument()

    await user.click(screen.getByLabelText('Очистить'))
    await user.type(screen.getByLabelText('Что вам нужно?'), 'молоко')
    await user.click(await screen.findByRole('button', { name: /Молоко/ }))
    expect(onAnswer).toHaveBeenLastCalledWith(expect.objectContaining({ type: 'product', target: 'milk' }))
  })

  it('если один из поисков упал, подсказок нет, но ничего не ломается', async () => {
    const user = userEvent.setup()
    mockFetch({ ['GET /api/medicines/suggest?q=' + encodeURIComponent('пара')]: () => jsonResponse([]) })
    render(<AskBox onAnswer={() => {}} />)
    await user.type(screen.getByLabelText('Что вам нужно?'), 'пара')
    await new Promise((r) => setTimeout(r, 500))
    expect(document.querySelector('.suggestions')).toBeNull()
    expect(screen.getByLabelText('Что вам нужно?')).toHaveValue('пара')
  })

  it('ограничивает длину ввода', () => {
    render(<AskBox onAnswer={() => {}} />)
    expect(screen.getByLabelText('Что вам нужно?')).toHaveAttribute('maxLength', '500')
  })
})
