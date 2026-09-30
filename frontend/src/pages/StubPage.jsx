import Icon from '../components/Icon'

function StubPage({ title, icon, color, onBack }) {
  return (
    <div className="page">
      {onBack && (
        <button type="button" className="back-btn" onClick={onBack}>
          <Icon name="back" size={20} />
          Назад
        </button>
      )}
      <h1 className="page-title">{title}</h1>
      <div className="card stub">
        <span className={'stub-icon color-' + (color || 'blue')}>
          <Icon name={icon} />
        </span>
        <p className="stub-text">Раздел скоро появится</p>
      </div>
    </div>
  )
}

export default StubPage
