let loading = null

export function loadYandexMaps() {
  if (window.ymaps) {
    return new Promise((resolve) => window.ymaps.ready(() => resolve(window.ymaps)))
  }

  if (!loading) {
    loading = new Promise((resolve, reject) => {
      const key = import.meta.env.VITE_YANDEX_MAPS_KEY
      const script = document.createElement('script')
      script.src = 'https://api-maps.yandex.ru/2.1/?lang=ru_RU' + (key ? '&apikey=' + key : '')
      script.onload = () => window.ymaps.ready(() => resolve(window.ymaps))
      script.onerror = () => {
        loading = null
        script.remove()
        reject(new Error('Не удалось загрузить карту'))
      }
      document.head.appendChild(script)
    })
  }
  return loading
}
