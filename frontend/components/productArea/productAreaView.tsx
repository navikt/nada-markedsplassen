import { System } from '@navikt/ds-icons'
import { Heading } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { PAItem, PAItems } from '../../pages/productArea/[id]'
import ProductAreaContent from './content'
import ProductAreaMobileMenu from './productAreaMobileMenu'
import ProductAreaSidebar from './sidebar'

interface ProductAreaViewProps {
  paItems: PAItems
  productAreas: any[]
}

const ProductAreaView = ({ paItems, productAreas }: ProductAreaViewProps) => {
  const router = useRouter()
  const currentItemName = (router.query.team as string) || paItems.length && paItems[0].name || ''
  const teamIdx = paItems.findIndex(
    (it) => it.name.toLowerCase() === currentItemName.toLowerCase() || it.id === currentItemName
  )

  const currentItem = teamIdx > 0 ? teamIdx : 0

  const initialTab = (item: PAItem) => {
    return item?.dashboardURL ? 'dashboard' : 'stories'
  }

  const [currentTab, setCurrentTab] = useState(initialTab(paItems[teamIdx]))
  const [open, setOpen] = useState(false)

  if(!currentItemName){
    return <div>Ingen produktområder samsvarer med søkekriteriene</div>
  }

  const pathComponents = router.asPath.split('?')[0].split('/')
  const currentProductAreaId = pathComponents[pathComponents.length - 1]

  const buildQueryString = (teamIdx: number) => {
    const queryString = new URLSearchParams()
    //team 0 is the product area
    if (teamIdx > 0) {
      queryString.append('team', paItems[teamIdx].name)
      return '?' + queryString.toString()
    }
    return ''
  }

  const handleSetCurrentItem = (newCurrent: number) => {
    const path =
      `/productArea/${currentProductAreaId}` + buildQueryString(newCurrent)
    router.push(path)
  }

  const handleSelectProductArea = (newProductAreaID: string) => {
    router.push(`/productArea/${newProductAreaID}`)
  }
  return (
    <div className="flex flex-col md:flex-row gap-3 py-4">
      <ProductAreaSidebar
          productAreaItems={paItems}
          setCurrentItem={handleSetCurrentItem}
          currentItem={currentItem}
          productAreas={productAreas}
          selectProductArea={handleSelectProductArea}
      />
      <div className="flex gap-4 items-center md:hidden">
          <Heading level="1" size="large">
          {paItems[currentItem].name}
          </Heading>
          <a
          className="flex gap-2 items-center"
          href="#"
          onClick={() => setOpen(!open)}
          >
          <System className="w-4 h-4" />
          Utforsk
          </a>
      </div>
      <ProductAreaMobileMenu
          open={open}
          setOpen={setOpen}
          productAreaItems={paItems}
          setCurrentItem={handleSetCurrentItem}
          productAreas={productAreas}
          selectProductArea={handleSelectProductArea}
      />
      <ProductAreaContent
          currentItem={paItems[currentItem]}
          currentTab={currentTab}
          setCurrentTab={setCurrentTab}
      />
    </div>
  )
}

export default ProductAreaView
