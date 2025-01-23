import { ExternalLinkIcon } from '@navikt/aksel-icons'
import { Link } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import * as React from 'react'
import humanizeDate from '../../lib/humanizeDate'
import { isEmail } from '../../lib/validators'
import { Subject, SubjectHeader } from '../subject'

interface DataproductDetailProps {
  product: any
  isOwner: boolean
  menuItems: Array<{
    title: string
    slug: string
    component: any
  }>
  currentPage: number
}

export const DataproductSidebar = ({
  product,
  isOwner,
  menuItems,
  currentPage,
}: DataproductDetailProps) => {
  const router = useRouter()

  const handleChange = (event: React.SyntheticEvent, newSlug: string) => {
    router.push(`/dataproduct/${product.id}/${product.slug}/${newSlug}`)
  }

  const makeTeamContactHref = (teamContact: string) => {
    return isEmail(teamContact)
      ? `mailto:${teamContact}`
      : `https://slack.com/app_redirect?channel=${teamContact}`
  }

  const truncate = (teamContact: string) => {
    return teamContact.length > 25
      ? teamContact.substring(0, 25) + "..."
      : teamContact
  }

  return (
    <div className="hidden md:flex md:flex-col md:gap-8 text-base pt-8 w-64 min-h-[calc(100vh-181px)]">
      <div className="flex w-64 flex-col gap-2">
        {menuItems.map(({ title, slug }, idx) =>
          currentPage == idx ? (
            typeof title === "string" 
            ? <p
              className="border-l-8 border-l-border-selected py-1 px-2 font-semibold"
              key={idx}
              dangerouslySetInnerHTML={{__html: title.replaceAll("_", "_<wbr>")}}
            />
            : <p
              className="border-l-8 border-l-border-selected py-1 px-2 font-semibold"
              key={idx}
            >{title}</p>
          ) : (
            // title might be a ReactNode (legg til datasett, lol), and we can't run .replace() on such an element
            typeof title === "string" 
            ? <a
              className="border-l-8 border-l-transparent font-semibold no-underline hover:underline hover:cursor-pointer py-1 px-2"
              href={`/dataproduct/${product.id}/${product.slug}/${slug}`}
              key={idx}
              dangerouslySetInnerHTML={{__html: title.replaceAll("_", "_<wbr>")}}
            />
            : <a
              className="border-l-8 border-l-transparent font-semibold no-underline hover:underline hover:cursor-pointer py-1 px-2"
              href={`/dataproduct/${product.id}/${product.slug}/${slug}`}
              key={idx}
            >
              {title}
            </a>
          )
        )}
      </div>
      <hr className="border-border-subtle mr-6" />
      <div className="h-fit w-64 text-base leading-4 pr-4 pb-0">
        <Subject>
          <Link
              href={`/productArea/${product.owner.productAreaID}?team=${product.owner.teamID}`}
              target="_blank"
              rel="noreferrer"
          >
            Utforsk alle teamets produkter
          </Link>
        </Subject>
        <SubjectHeader>Ansvarlig team (Teamkatalogen)</SubjectHeader>
        <Subject>
          {product.owner?.teamID && product.owner.teamkatalogenURL? (
            <Link
              href={product.owner.teamkatalogenURL}
              target="_blank"
              rel="noreferrer"
            >
              {product.owner.group.split('@')[0]} <ExternalLinkIcon />
            </Link>
          ) : (
            product.owner?.group.split('@')[0]
          )}
        </Subject>
        <SubjectHeader>
        {product.owner?.teamContact && isEmail(product.owner.teamContact) ? 'Kontaktpunkt (e-post)' : 'Kontaktpunkt (Slack)'}
          </SubjectHeader>
        <Subject>
          {product.owner?.teamContact ? (
            <Link
              href={makeTeamContactHref(product.owner.teamContact)}
              target="_blank"
              rel="noreferrer"
              aria-label={product.owner.teamContact}
              title={product.owner.teamContact}
            >
              {truncate(product.owner.teamContact)} <ExternalLinkIcon />
            </Link>
          ) : (
            'Ukjent'
          )}
        </Subject>

        <SubjectHeader>Opprettet</SubjectHeader>
        <Subject>{humanizeDate(product.created)}</Subject>
      </div>
    </div>
  )
}
