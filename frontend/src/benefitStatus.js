export const statusNames = {
  not_started: 'Не начато',
  collecting: 'Собираю документы',
  submitted: 'Заявление подано',
  review: 'На рассмотрении',
  approved: 'Одобрено',
  rejected: 'Отказ',
}

export const categoryIcons = {
  Документы: 'document',
  Пенсия: 'wallet',
  Льготы: 'percent',
  Досуг: 'heart',
}

// шаги полоски статуса: решение — последний шаг и для одобрения, и для отказа
export const statusSteps = ['collecting', 'submitted', 'review', 'decision']

export function stepIndex(status) {
  if (status === 'approved' || status === 'rejected') {
    return 3
  }
  return statusSteps.indexOf(status)
}
