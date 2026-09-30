import { describe, expect, it } from 'vitest'
import { ageFrom, formatDistance, formatFullDate, formatRating, formatShortDate, phoneLink, reviewsText, routeLink, yearsText } from './format'

describe('format', () => {
  it('расстояние', () => {
    expect(formatDistance(0)).toBe('10 м')
    expect(formatDistance(0.004)).toBe('10 м')
    expect(formatDistance(0.53)).toBe('530 м')
    expect(formatDistance(0.999)).toBe('1000 м')
    expect(formatDistance(1)).toBe('1,0 км')
    expect(formatDistance(1.25)).toBe('1,3 км')
    expect(formatDistance(12.04)).toBe('12,0 км')
  })

  it('рейтинг', () => {
    expect(formatRating(4.8)).toBe('4,8')
    expect(formatRating(5)).toBe('5,0')
  })

  it('склонение отзывов', () => {
    expect(reviewsText(0)).toBe('0 отзывов')
    expect(reviewsText(1)).toBe('1 отзыв')
    expect(reviewsText(3)).toBe('3 отзыва')
    expect(reviewsText(5)).toBe('5 отзывов')
    expect(reviewsText(11)).toBe('11 отзывов')
    expect(reviewsText(12)).toBe('12 отзывов')
    expect(reviewsText(21)).toBe('21 отзыв')
    expect(reviewsText(104)).toBe('104 отзыва')
    expect(reviewsText(111)).toBe('111 отзывов')
  })

  it('ссылка на маршрут', () => {
    expect(routeLink(55.75, 37.61)).toBe('https://yandex.ru/maps/?rtext=~55.75,37.61&rtt=pd')
  })

  it('даты', () => {
    expect(formatShortDate('2026-10-03')).toBe('3 октября')
    expect(formatShortDate('2026-01-31')).toBe('31 января')
    expect(formatFullDate('1956-03-12')).toBe('12 марта 1956')
  })

  it('возраст и склонение лет', () => {
    const today = new Date(2026, 8, 30)
    expect(ageFrom('1956-03-12', today)).toBe(70)
    expect(ageFrom('1981-10-05', today)).toBe(44)
    expect(ageFrom('1981-09-30', today)).toBe(45)
    expect(yearsText(1)).toBe('1 год')
    expect(yearsText(44)).toBe('44 года')
    expect(yearsText(45)).toBe('45 лет')
    expect(yearsText(71)).toBe('71 год')
    expect(yearsText(111)).toBe('111 лет')
  })

  it('ссылка на звонок', () => {
    expect(phoneLink('+7 (495) 123-45-67')).toBe('tel:+74951234567')
    expect(phoneLink('8 843 111 22 33')).toBe('tel:88431112233')
    expect(phoneLink('122')).toBe('tel:122')
  })
})
