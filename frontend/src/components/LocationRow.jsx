import Icon from './Icon'

function LocationRow({ location, onOpenProfile }) {
  return (
    <div className="location-row">
      <Icon name="pin" size={18} />
      <span className="location-text">
        {location.isDefault ? 'Адрес не указан, показываем центр Москвы' : 'Рядом с: ' + location.address}
      </span>
      <button type="button" className="link-btn location-btn" onClick={onOpenProfile}>
        {location.isDefault ? 'Указать' : 'Изменить'}
      </button>
    </div>
  )
}

export default LocationRow
