import AskBox from '../components/AskBox'
import Features from '../components/Features'
import ForYou from '../components/ForYou'
import TodayTasks from '../components/TodayTasks'

function HomePage({ onOpenTab, onOpenFeature, onOpenMedicine, onOpenProduct, onOpenShop, onOpenProfile }) {
  function handleAnswer(answer) {
    if (answer.type === 'tab') {
      onOpenTab(answer.target)
    } else if (answer.type === 'feature' && answer.feature) {
      onOpenFeature(answer.feature)
    } else if (answer.type === 'medicine' && answer.medicine) {
      onOpenMedicine(answer.medicine)
    } else if (answer.type === 'product' && answer.product) {
      onOpenProduct(answer.product)
    }
  }

  return (
    <>
      <AskBox onAnswer={handleAnswer} onOpenProfile={onOpenProfile} />
      <div className="page">
        <ForYou onOpenShop={onOpenShop} />
        <TodayTasks />
        <Features onOpen={onOpenFeature} />
      </div>
    </>
  )
}

export default HomePage
