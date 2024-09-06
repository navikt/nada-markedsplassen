import { Heading, Link } from '@navikt/ds-react'
import ExploreAreasIcon from '../lib/icons/exploreAreasIcon'
import { useGetProductAreas } from '../../lib/rest/productAreas'

const ProductAreaHasItems = (p: any)=> !!p?.teams.filter((it: any)=> it.dataproductsNumber+ it.storiesNumber> 0).length

const ProductAreaLinks = () => {
  var defaultProductAreaID = '6b149078-927b-4570-a1ce-97bbb9499fb6'
  const {productAreas, error} = useGetProductAreas()
  const productAreaWithItems = productAreas?.filter(it=> ProductAreaHasItems(it));
  if(productAreaWithItems.length>0){
    defaultProductAreaID = productAreaWithItems.find(it=> it.id== defaultProductAreaID)?.id || productAreaWithItems[0].id
  }else if(error){
    return null
  }

  return (
    <div className="border border-border-default bg-white rounded-lg w-11/12 md:w-[17rem] md:h-[22rem] p-4 pt-8 flex items-center flex-col gap-8">
      <ExploreAreasIcon />
      <div>
        <Heading level="2" size="small">
          <Link
            href={`/productArea/${defaultProductAreaID}`}
          >
            Utforsk områder
          </Link>
        </Heading>
        <p>Er du nysgjerrig på hva de ulike områdene i NAV gjør?
          Her kan du utforske dataprodukter og fortellinger noen av områdene har publisert.</p>
      </div>
    </div>
  )
}

export default ProductAreaLinks
