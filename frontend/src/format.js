export function formatDistance(km) {
  if (km < 1) {
    return Math.max(10, Math.round(km * 100) * 10) + ' м'
  }
  return km.toFixed(1).replace('.', ',') + ' км'
}

export function formatRating(rating) {
  return rating.toFixed(1).replace('.', ',')
}

export function reviewsText(count) {
  const lastTwo = count % 100
  const last = count % 10
  if (lastTwo >= 11 && lastTwo <= 14) {
    return count + ' отзывов'
  }
  if (last === 1) {
    return count + ' отзыв'
  }
  if (last >= 2 && last <= 4) {
    return count + ' отзыва'
  }
  return count + ' отзывов'
}

export function routeLink(lat, lon) {
  return 'https://yandex.ru/maps/?rtext=~' + lat + ',' + lon + '&rtt=pd'
}

const monthsGenitive = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
  'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']

export function formatShortDate(value) {
  const parts = value.split('-')
  return Number(parts[2]) + ' ' + monthsGenitive[Number(parts[1]) - 1]
}

export function formatFullDate(value) {
  return formatShortDate(value) + ' ' + value.split('-')[0]
}

export function yearsText(n) {
  const lastTwo = n % 100
  const last = n % 10
  if (lastTwo >= 11 && lastTwo <= 14) {
    return n + ' лет'
  }
  if (last === 1) {
    return n + ' год'
  }
  if (last >= 2 && last <= 4) {
    return n + ' года'
  }
  return n + ' лет'
}

export function ageFrom(value, today) {
  const parts = value.split('-').map(Number)
  let age = today.getFullYear() - parts[0]
  const month = today.getMonth() + 1
  if (month < parts[1] || (month === parts[1] && today.getDate() < parts[2])) {
    age = age - 1
  }
  return age
}

export function phoneLink(phone) {
  return 'tel:' + phone.replace(/[^\d+]/g, '')
}

function pad(n) {
  return String(n).padStart(2, '0')
}

// в аптеку предлагаем пойти в ближайший целый час, а поздно вечером — завтра в 10:00
export function suggestVisit(now) {
  const next = now.getHours() + 1
  if (next >= 21) {
    const tomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1)
    return { date: dateString(tomorrow), time: '10:00', tomorrow: true }
  }
  return { date: dateString(now), time: pad(Math.max(next, 8)) + ':00', tomorrow: false }
}

function dateString(d) {
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
}
