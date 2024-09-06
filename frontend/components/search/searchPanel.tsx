import { Accordion, Search } from '@navikt/ds-react'
import { useState } from 'react'
import { FilterType, SearchParam } from '../../pages/search'
import { FiltersPicker, FilterTreeNode } from './filtersPicker'

export interface SearchPanelProps {
  searchParam: SearchParam
  filtersTree: Map<FilterType, FilterTreeNode>
  updateQuery: (searchParam: SearchParam) => void
}

export const SearchPanel = ({
  searchParam,
  filtersTree,
  updateQuery,
}: SearchPanelProps) => {
  const [updatedSearchTerm, setUpdatedSearchTerm] = useState(
    searchParam.freeText || ''
  )
  const onToggleTeam = (filter: string) => {
    const picked = !searchParam.teams?.find((it) => filter === it)
    const updatedFilterValues = picked
      ? [...new Set([...(searchParam.teams || []), filter])]
      : searchParam.teams?.filter((it) => it != filter)
    updateQuery({ ...searchParam, teams: updatedFilterValues })
  }

  const onPickKeyword = (filter: string) => {
    const picked = !searchParam.keywords?.find((it) => filter === it)
    const updatedFilterValues = picked
      ? [...new Set([...(searchParam.keywords || []), filter]).values()]
      : searchParam.keywords?.filter((it) => it != filter)
    updateQuery({ ...searchParam, keywords: updatedFilterValues })
  }

  const onToggleFilter = (filterTree: FilterType, filter: string) => {
    switch (filterTree) {
      case 'Områder':
        onToggleTeam(filter)
        break
      case 'Nøkkelord':
        onPickKeyword(filter)
        break
      default:
        console.log(`Unrecognized filter tree -> ${filterTree}`)
        break
    }
  }

  return (
    <div>
      <form
        className="self-center px-5 mb-5 mt-3"
        onSubmit={(e) => {
          e.preventDefault()
          updateQuery({ ...searchParam, freeText: updatedSearchTerm })
        }}
      >
        <Search
          label="panel søk"
          size="small"
          variant="simple"
          placeholder="Søk"
          value={updatedSearchTerm}
          onChange={(text) => {
            setUpdatedSearchTerm(text)
          }}
        />
      </form>
      <Accordion>
        {Array.from(filtersTree.entries()).map((tree) => (
          <FiltersPicker
            key={tree[0]}
            header={tree[0]}
            filtersTree={tree[1]}
            onToggle={(filter) => onToggleFilter(tree[0], filter)}
          ></FiltersPicker>
        ))}
      </Accordion>
    </div>
  )
}
