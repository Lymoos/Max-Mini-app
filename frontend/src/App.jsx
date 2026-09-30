import { useEffect, useState } from 'react'
import BottomNav from './components/BottomNav'
import BenefitPage from './pages/BenefitPage'
import ClinicPage from './pages/ClinicPage'
import ClinicsPage from './pages/ClinicsPage'
import DoctorPage from './pages/DoctorPage'
import DocumentsPage from './pages/DocumentsPage'
import GoodsPage from './pages/GoodsPage'
import GuidePage from './pages/GuidePage'
import HelpPage from './pages/HelpPage'
import HomePage from './pages/HomePage'
import MedicinesPage from './pages/MedicinesPage'
import ProfilePage from './pages/ProfilePage'
import ShopPage from './pages/ShopPage'
import SocialPage from './pages/SocialPage'
import StubPage from './pages/StubPage'
import { openLinksThroughMax, openStartScreen, startParam } from './startParam'

function App() {
  const [tab, setTab] = useState('home')
  const [feature, setFeature] = useState(null)
  const [medicine, setMedicine] = useState(null)
  const [product, setProduct] = useState(null)
  const [shopId, setShopId] = useState(null)
  const [clinicId, setClinicId] = useState(null)
  const [doctorId, setDoctorId] = useState(null)
  const [benefitId, setBenefitId] = useState(null)
  const [guideId, setGuideId] = useState(null)
  const [specialtyId, setSpecialtyId] = useState('')
  const [autoOpenClinic, setAutoOpenClinic] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const [profileVersion, setProfileVersion] = useState(0)

  useEffect(() => {
    const param = startParam()
    if (param !== '') {
      openStartScreen(param, {
        tab: openTab,
        feature: openFeature,
        medicine: openMedicine,
        product: openProduct,
        benefit: openBenefit,
        doctor: openDoctor,
        guide: openGuide,
        profile: () => setProfileOpen(true),
      })
    }
    document.addEventListener('click', openLinksThroughMax)
    return () => document.removeEventListener('click', openLinksThroughMax)
  }, [])

  function openTab(name) {
    setFeature(null)
    setMedicine(null)
    setProduct(null)
    setShopId(null)
    setClinicId(null)
    setDoctorId(null)
    setBenefitId(null)
    setGuideId(null)
    setTab(name)
    window.scrollTo(0, 0)
  }

  function openFeature(f) {
    if (f.id === 'pharmacy') {
      openTab('medicines')
      return
    }
    setFeature(f)
    setProduct(null)
    setShopId(null)
    setClinicId(null)
    setDoctorId(null)
    setSpecialtyId('')
    setAutoOpenClinic(false)
    window.scrollTo(0, 0)
  }

  // из поиска «нужен кардиолог»: сразу открываем свою поликлинику с врачами этой специальности
  function openDoctor(id) {
    openFeature({ id: 'doctor', title: 'Запись к врачу' })
    setSpecialtyId(id)
    setAutoOpenClinic(id !== '')
  }

  function openGuide(id) {
    openTab('help')
    setGuideId(id)
  }

  function openMedicine(m) {
    openTab('medicines')
    setMedicine(m)
  }

  function openProduct(p) {
    setFeature({ id: 'goods', title: 'Товары рядом' })
    setProduct(p)
    setShopId(null)
    window.scrollTo(0, 0)
  }

  function openShop(shop) {
    setShopId(shop.id)
    window.scrollTo(0, 0)
  }

  function openBenefit(id) {
    openTab('documents')
    setBenefitId(id)
  }

  function closeFeature() {
    setFeature(null)
    setProduct(null)
    setClinicId(null)
    setDoctorId(null)
    setSpecialtyId('')
    setAutoOpenClinic(false)
  }

  let page
  if (feature && feature.id === 'doctor') {
    if (doctorId !== null) {
      page = <DoctorPage key={doctorId} doctorId={doctorId} onBack={() => setDoctorId(null)} />
    } else if (clinicId !== null) {
      page = (
        <ClinicPage
          key={clinicId}
          clinicId={clinicId}
          specialtyId={specialtyId}
          onOpenDoctor={(id) => {
            setDoctorId(id)
            window.scrollTo(0, 0)
          }}
          onBack={() => setClinicId(null)}
        />
      )
    } else {
      page = (
        <ClinicsPage
          key={profileVersion}
          specialtyId={specialtyId}
          autoOpen={autoOpenClinic}
          onOpenClinic={(id) => {
            setAutoOpenClinic(false)
            setClinicId(id)
            window.scrollTo(0, 0)
          }}
          onBack={closeFeature}
        />
      )
    }
  } else if (feature && feature.id === 'social') {
    page = <SocialPage key={profileVersion} onBack={closeFeature} onOpenProfile={() => setProfileOpen(true)} />
  } else if (shopId !== null) {
    page = <ShopPage key={shopId} shopId={shopId} onBack={() => setShopId(null)} />
  } else if (feature && feature.id === 'goods') {
    page = (
      <GoodsPage
        key={profileVersion}
        product={product}
        onPickProduct={openProduct}
        onClearProduct={() => setProduct(null)}
        onOpenShop={openShop}
        onOpenProfile={() => setProfileOpen(true)}
        onBack={closeFeature}
      />
    )
  } else if (feature) {
    page = <StubPage title={feature.title} icon={feature.icon} color={feature.color} onBack={() => setFeature(null)} />
  } else if (tab === 'home') {
    page = (
      <HomePage
        key={profileVersion}
        onOpenTab={openTab}
        onOpenFeature={openFeature}
        onOpenMedicine={openMedicine}
        onOpenProduct={openProduct}
        onOpenShop={openShop}
        onOpenBenefit={openBenefit}
        onOpenDoctor={openDoctor}
        onOpenGuide={openGuide}
        onOpenProfile={() => setProfileOpen(true)}
      />
    )
  } else if (tab === 'medicines') {
    page = (
      <MedicinesPage
        key={profileVersion}
        medicine={medicine}
        onPickMedicine={openMedicine}
        onClearMedicine={() => setMedicine(null)}
        onOpenProfile={() => setProfileOpen(true)}
      />
    )
  } else if (tab === 'documents') {
    page =
      benefitId !== null ? (
        <BenefitPage key={benefitId} benefitId={benefitId} onBack={() => setBenefitId(null)} />
      ) : (
        <DocumentsPage
          key={profileVersion}
          onOpenBenefit={(id) => {
            setBenefitId(id)
            window.scrollTo(0, 0)
          }}
        />
      )
  } else {
    page =
      guideId !== null ? (
        <GuidePage key={guideId} guideId={guideId} onBack={() => setGuideId(null)} />
      ) : (
        <HelpPage
          onOpenGuide={(id) => {
            setGuideId(id)
            window.scrollTo(0, 0)
          }}
        />
      )
  }

  return (
    <div className="app">
      <main>{page}</main>
      <BottomNav active={tab} onChange={openTab} />
      {profileOpen && (
        <ProfilePage onClose={() => setProfileOpen(false)} onSaved={() => setProfileVersion((v) => v + 1)} />
      )}
    </div>
  )
}

export default App
