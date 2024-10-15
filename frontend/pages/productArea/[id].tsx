import { Loader } from '@navikt/ds-react'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { useEffect } from 'react'
import ErrorMessage from '../../components/lib/error'
import InnerContainer from '../../components/lib/innerContainer'
import ProductAreaView from '../../components/productArea/productAreaView'
import amplitudeLog from '../../lib/amplitude'
import { ProductAreaWithAssets } from '../../lib/rest/generatedDto'
import { useGetProductArea, useGetProductAreas } from '../../lib/rest/productAreas'


export interface PAItem {
  name: string
  id?: string
  dashboardURL?: string
  dataproducts: {
    __typename?: 'Dataproduct'
    id: string
    name: string
    description: string
    created: any
    lastModified: any
    keywords: Array<string>
    slug: string
    owner: {
      __typename?: 'Owner'
      group: string
      teamkatalogenURL?: string | null | undefined
      teamContact?: string | null | undefined
    }
  }[]
  stories: {
    __typename?: 'Story'
    id: string
    name: string
    description: string
    created: any
    keywords: Array<string>
    lastModified?: any | null | undefined
  }[]
  insightProducts: {
    __typename?: 'InsightProduct'
    id: string
    name: string
    link: string
    type: string
    description: string
    created: any
    lastModified?: any
    keywords: Array<string>
  }[]
}

export interface PAItems extends Array<PAItem> { }

const isRelevantPA = (productArea: any) => {
  return (
    productArea.dataproducts.length
    || productArea.stories.length
    || productArea.insightProducts.length
  )
}

const createPAItems = (productArea: ProductAreaWithAssets) => {
  let items: any[] = []
  if (isRelevantPA(productArea)) {
    productArea.teamsWithAssets
      .slice()
      .filter(
        (it: any) =>
          it.dataproducts.length > 0 ||
          it.stories.length > 0 ||
          it.insightProducts.length > 0
      )
      .sort((a: any, b: any) => {
        return (
          b.dataproducts.length +
          b.insightProducts.length -
          (a.dataproducts.length + a.stories.length + a.insightProducts.length)
        )
      })
      .forEach((t: any) => {
        items.push({
          name: t.name,
          id: t.id,
          dashboardURL: t.dashboardURL,
          dataproducts: t.dataproducts,
          stories: t.stories,
          insightProducts: t.insightProducts,
        })
      })

      items.unshift(
        items.reduce((paassets, item)=>{
          paassets.dataproducts.push(...item.dataproducts)
          paassets.stories.push(...item.stories)
          paassets.insightProducts.push(...item.insightProducts)
          return paassets
        }, {
          name: productArea.name,
          dashboardURL: productArea.dashboardURL,
          id: productArea.id,
          dataproducts: [],
          stories: [],
          insightProducts: [],
        }))
  }

  return items
}

interface ProductAreaProps {
  id: string
  productAreas: any[]
}

const ProductArea = ({ id, productAreas }: ProductAreaProps) => {
  const {productArea, loading, error} = useGetProductArea(id)

  useEffect(() => {
    if (!loading && productArea) {
      const eventProperties = {
        sidetittel: 'poside',
        title: productArea.name,
      }
      amplitudeLog('sidevisning', eventProperties)
    }
  })

  if (error)
    return <ErrorMessage error={error} />
  if (loading || !productArea)
    return <Loader />

  const paItems = createPAItems(productArea)

  return <ProductAreaView paItems={paItems} productAreas={productAreas} />
}

const ProductAreaPage = () => {
  const router = useRouter()
  const { productAreas, loading, error } = useGetProductAreas()

  if (!router.isReady) return <Loader />

  if (error)
    return (
      <div>
        <p>Teamkatalogen er utilgjengelig. Se på slack channel <Link href="https://slack.com/app_redirect?channel=teamkatalogen">#Teamkatalogen</Link></p>
        <ErrorMessage error={error} />
      </div>
    )
  if ( loading ||!productAreas?.length)
    return <Loader />

  return (
    <InnerContainer>
      <ProductArea
        id={router.query.id as string}
        productAreas={productAreas}
      />
    </InnerContainer>
  )
}

export default ProductAreaPage
