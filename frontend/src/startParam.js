// бот открывает миниапп кнопкой с данными вида med_paracetamol — сразу показываем нужный экран
export function startParam() {
  const unsafe = window.WebApp && window.WebApp.initDataUnsafe
  if (!unsafe || !unsafe.start_param) {
    return ''
  }
  return String(unsafe.start_param)
}

export function openStartScreen(param, open) {
  if (param.startsWith('med_')) {
    open.medicine({ id: param.slice(4), name: '', form: '' })
  } else if (param.startsWith('prod_')) {
    open.product({ id: param.slice(5), name: '', unit: '' })
  } else if (param.startsWith('benefit_')) {
    open.benefit(param.slice(8))
  } else if (param.startsWith('doctor_')) {
    open.doctor(param.slice(7))
  } else if (param.startsWith('guide_')) {
    open.guide(param.slice(6))
  } else if (param === 'profile') {
    open.profile()
  } else if (param === 'pharmacy' || param === 'medicines') {
    open.tab('medicines')
  } else if (param === 'goods' || param === 'doctor' || param === 'social') {
    open.feature({ id: param })
  } else if (param === 'documents' || param === 'help') {
    open.tab(param)
  }
}

// внешние ссылки внутри MAX нужно открывать через мост, иначе webview может их не открыть
export function openLinksThroughMax(event) {
  const link = event.target.closest && event.target.closest('a[target="_blank"]')
  if (link && window.WebApp && window.WebApp.openLink) {
    event.preventDefault()
    window.WebApp.openLink(link.href)
  }
}
