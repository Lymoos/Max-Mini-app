import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { demoFeatures, jsonResponse, mockFetch } from '../testUtils'
import Features from './Features'

describe('Features', () => {
  it('рисует плитки и открывает нужную по нажатию', async () => {
    const user = userEvent.setup()
    mockFetch({ 'GET /api/features': () => jsonResponse(demoFeatures) })
    const onOpen = vi.fn()
    render(<Features onOpen={onOpen} />)

    await user.click(await screen.findByRole('button', { name: 'Товары рядом' }))
    expect(onOpen).toHaveBeenCalledWith(demoFeatures[1])
    expect(screen.getAllByRole('button')).toHaveLength(2)
  })

  it('пока грузится — четыре серых плитки-силуэта, нажать их нельзя', async () => {
    mockFetch({ 'GET /api/features': () => jsonResponse(demoFeatures) })
    render(<Features onOpen={() => {}} />)

    expect(document.querySelectorAll('.skeleton-tile')).toHaveLength(4)
    expect(screen.queryAllByRole('button')).toHaveLength(0)
    await screen.findByRole('button', { name: 'Товары рядом' })
    expect(document.querySelector('.skeleton-tile')).toBeNull()
  })

  it('ошибка и повтор', async () => {
    const user = userEvent.setup()
    mockFetch({})
    render(<Features onOpen={() => {}} />)

    expect(await screen.findByText('Не удалось загрузить')).toBeInTheDocument()

    mockFetch({ 'GET /api/features': () => jsonResponse(demoFeatures) })
    await user.click(screen.getByText('Повторить'))
    expect(await screen.findByText('Аптеки рядом')).toBeInTheDocument()
  })
})
