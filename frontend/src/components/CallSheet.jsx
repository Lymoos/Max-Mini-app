import { phoneLink } from '../format'
import Icon from './Icon'

function CallSheet({ place, onClose }) {
  return (
    <div className="overlay overlay-top" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Позвонить">
        <p className="call-name">{place.name}</p>
        <p className="call-phone">{place.phone}</p>
        <a className="btn btn-primary call-btn" href={phoneLink(place.phone)}>
          <Icon name="phone" size={20} />
          Позвонить
        </a>
        <button type="button" className="btn btn-secondary call-cancel" onClick={onClose}>
          Отмена
        </button>
      </div>
    </div>
  )
}

export default CallSheet
