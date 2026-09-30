function PlatformSwitch({ value, onChange }) {
  return (
    <div className="segmented" role="group" aria-label="Мой телефон">
      <span className={value === 'iphone' ? 'segmented-thumb segmented-thumb-right' : 'segmented-thumb'} />
      <button type="button" className="segmented-btn" onClick={() => onChange('android')} aria-pressed={value === 'android'}>
        Android
      </button>
      <button type="button" className="segmented-btn" onClick={() => onChange('iphone')} aria-pressed={value === 'iphone'}>
        iPhone
      </button>
    </div>
  )
}

export default PlatformSwitch
