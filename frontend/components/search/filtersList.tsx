import { Error } from '@navikt/ds-icons'
import * as React from 'react'
import { emptyFilter, isEmptyFilter, SearchParam } from '../../pages/search'

const FilterRow = ({ children }: { children: React.ReactNode }) => {
  return <div className="flex gap-2">{children}</div>
}

const FilterPill = ({
  all,
  children,
  onClick,
  className,
}: {
  all?: boolean
  children: React.ReactNode
  onClick: () => void
  className?: string
}) => {
  return (
    <span
      onClick={onClick}
      className={`${
        className || ''
      } svg-scale flex items-center gap-1 cursor-pointer text-xs p-2
      ${
        all
          ? 'bg-surface-action text-text-on-inverted rounded-sm'
          : 'bg-gray-100 rounded-3xl'
      }
      ${all ? 'hover:bg-surface-action-hover' : 'hover:bg-gray-300'}`}
    >
      {children}
    </span>
  )
}

interface FiltersListProps {
  searchParam: SearchParam
  updateQuery: (updatedParam: SearchParam) => void
  className?: string
}

const FiltersList = ({
  searchParam,
  updateQuery,
  className,
}: FiltersListProps) => (
  <div className={className}>
    {!isEmptyFilter(searchParam) && (
      <FilterRow>
        <FilterPill
          all={true}
          onClick={() =>
            updateQuery({
              ...emptyFilter,
              preferredType: searchParam.preferredType,
            })
          }
        >
          Fjern alle filtre
          <Error />
        </FilterPill>
        {!!searchParam.freeText && (
          <FilterPill
            key={'searchTerm'}
            onClick={() => updateQuery({ ...searchParam, freeText: '' })}
          >
            {searchParam.freeText}
            <Error />
          </FilterPill>
        )}
        {searchParam.teams?.map((t, index) => (
          <FilterPill
            key={`searchTeam${index}`}
            onClick={() =>
              updateQuery({
                ...searchParam,
                teams: searchParam.teams?.filter((st) => st !== t),
              })
            }
          >
            {t}
            <Error />
          </FilterPill>
        ))}
        {searchParam.keywords?.map((k, index) => (
          <FilterPill
            key={`searchKeyword${index}`}
            onClick={() =>
              updateQuery({
                ...searchParam,
                keywords: searchParam.keywords?.filter((sk) => sk !== k),
              })
            }
            className=""
          >
            {k.includes(" (") ? k.split(" (")[0] : k}
            <Error />
          </FilterPill>
        ))}
      </FilterRow>
    )}
  </div>
)

export default FiltersList
