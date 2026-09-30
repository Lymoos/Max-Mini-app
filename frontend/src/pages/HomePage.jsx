import { useState } from 'react'
import AskBox from '../components/AskBox'
import Features from '../components/Features'
import ForYou from '../components/ForYou'
import TodayTasks from '../components/TodayTasks'

function HomePage({
  onOpenTab,
  onOpenFeature,
  onOpenMedicine,
  onOpenProduct,
  onOpenShop,
  onOpenBenefit,
  onOpenDoctor,
  onOpenGuide,
  onOpenProfile,
}) {
  const [tasksVersion, setTasksVersion] = useState(0)

  function handleAnswer(answer) {
    if (answer.type === 'tab') {
      onOpenTab(answer.target)
    } else if (answer.type === 'feature' && answer.feature) {
      onOpenFeature(answer.feature)
    } else if (answer.type === 'medicine' && answer.medicine) {
      onOpenMedicine(answer.medicine)
    } else if (answer.type === 'product' && answer.product) {
      onOpenProduct(answer.product)
    } else if (answer.type === 'doctor') {
      onOpenDoctor(answer.target || '')
    } else if (answer.type === 'benefit') {
      onOpenBenefit(answer.target)
    } else if (answer.type === 'guide') {
      onOpenGuide(answer.target)
    } else if (answer.type === 'profile') {
      onOpenProfile()
    }
  }

  return (
    <>
      <AskBox onAnswer={handleAnswer} onOpenProfile={onOpenProfile} onTaskAdded={() => setTasksVersion((v) => v + 1)} />
      <div className="page">
        <TodayTasks key={tasksVersion} />
        <ForYou onOpenShop={onOpenShop} onOpenBenefit={onOpenBenefit} />
        <Features onOpen={onOpenFeature} />
      </div>
    </>
  )
}

export default HomePage
