async function request(path, options) {
  let res
  try {
    res = await fetch('/api' + path, options)
  } catch {
    throw new Error('Нет связи с сервером')
  }

  let data = null
  try {
    data = await res.json()
  } catch {
    data = null
  }

  if (!res.ok) {
    throw new Error(data && data.error ? data.error : 'Ошибка сервера')
  }
  return data
}

export function getTodayTasks() {
  return request('/tasks/today')
}

export function createTask(task) {
  return request('/tasks', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(task),
  })
}

export function setTaskDone(id, done) {
  return request('/tasks/' + id, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ done: done }),
  })
}

export function getFeatures() {
  return request('/features')
}

export function sendAsk(text) {
  return request('/ask', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text: text }),
  })
}

export function getProfile() {
  return request('/profile')
}

export function saveProfile(profile) {
  return request('/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(profile),
  })
}

export function searchAddress(query) {
  return request('/geo/search?q=' + encodeURIComponent(query))
}

export function reverseAddress(lat, lon) {
  return request('/geo/reverse?lat=' + lat + '&lon=' + lon)
}

export function suggestMedicines(query) {
  return request('/medicines/suggest?q=' + encodeURIComponent(query))
}

export function getOffers(medicineId) {
  return request('/medicines/' + encodeURIComponent(medicineId) + '/offers')
}

export function getNearbyPharmacies() {
  return request('/pharmacies/nearby')
}

export function saveAddress(place) {
  return request('/profile/address', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(place),
  })
}

export function getForYou() {
  return request('/for-you')
}

export function suggestProducts(query) {
  return request('/products/suggest?q=' + encodeURIComponent(query))
}

export function getProductOffers(productId) {
  return request('/products/' + encodeURIComponent(productId) + '/offers')
}

export function getNearbyShops() {
  return request('/shops/nearby')
}

export function getShop(id) {
  return request('/shops/' + id)
}

export function getClinics() {
  return request('/clinics')
}

export function getClinic(id) {
  return request('/clinics/' + id)
}

export function getDoctor(id) {
  return request('/doctors/' + id)
}

export function saveRegistration(place) {
  return request('/profile/registration', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(place),
  })
}

export function saveClinic(clinicId) {
  return request('/profile/clinic', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ clinicId: clinicId }),
  })
}

export function getSocialNearby() {
  return request('/social/nearby')
}

export function getBenefits() {
  return request('/benefits')
}

export function getBenefit(id) {
  return request('/benefits/' + encodeURIComponent(id))
}

export function saveBenefit(id, changes) {
  return request('/benefits/' + encodeURIComponent(id), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(changes),
  })
}

export function getGuides() {
  return request('/guides')
}

export function getGuide(id) {
  return request('/guides/' + encodeURIComponent(id))
}

export function markGuide(id, read) {
  return request('/guides/' + encodeURIComponent(id) + '/read', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ read: read }),
  })
}
