import Icon from './Icon'

const tabs = [
  { id: 'home', title: 'Дом', icon: 'home' },
  { id: 'medicines', title: 'Лекарства', icon: 'pill' },
  { id: 'documents', title: 'Документы', icon: 'document' },
  { id: 'help', title: 'Помощь', icon: 'help' },
]

function BottomNav({ active, onChange }) {
  return (
    <nav className="bottom-nav" aria-label="Основное меню">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          type="button"
          className={tab.id === active ? 'nav-item nav-item-active' : 'nav-item'}
          onClick={() => onChange(tab.id)}
          aria-current={tab.id === active ? 'page' : undefined}
        >
          <Icon name={tab.icon} />
          <span>{tab.title}</span>
        </button>
      ))}
    </nav>
  )
}

export default BottomNav
