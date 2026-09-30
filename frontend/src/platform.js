const KEY = 'phonePlatform'

export function detectPlatform() {
  try {
    const saved = localStorage.getItem(KEY)
    if (saved === 'android' || saved === 'iphone') {
      return saved
    }
  } catch {
    // хранилище недоступно — определяем по браузеру
  }
  return /iPhone|iPad|iPod/.test(navigator.userAgent) ? 'iphone' : 'android'
}

export function savePlatform(value) {
  try {
    localStorage.setItem(KEY, value)
  } catch {
    // не страшно, просто не запомним
  }
}
