import { Accordion, Checkbox } from '@navikt/ds-react'
import { FoldablePanel } from '../lib/foldablePanel/foldablePanel'

export interface FilterTreeNode {
  [s: string]: boolean | FilterTreeNode
}

export interface FiltersPickerProps {
  header: string
  filtersTree: FilterTreeNode
  onToggle: (filter: string) => void
  className?: string
}

export const FiltersPicker = ({
  header,
  filtersTree,
  onToggle,
  className,
}: FiltersPickerProps) => {
  return (
    <div className={className}>
      <Accordion.Item>
        <Accordion.Header>{header}</Accordion.Header>
        <Accordion.Content>
          <FiltersTreeView filtersTree={filtersTree} onToggle={onToggle} />
        </Accordion.Content>
      </Accordion.Item>
    </div>
  )
}

interface FilterTreeViewProps {
  filtersTree: FilterTreeNode | boolean
  onToggle: (filter: string) => void
  className?: string
}

const isChecked = (filtersTree: FilterTreeNode | boolean): boolean =>
  typeof filtersTree === 'boolean'
    ? (filtersTree as boolean)
    : Object.entries(filtersTree).every((it) => isChecked(it[1]))

const FiltersTreeView = ({
  className,
  onToggle,
  filtersTree,
}: FilterTreeViewProps) => {
  return (
    <div className={className}>
      {typeof filtersTree === 'object' &&
        !!filtersTree &&
        Object.getOwnPropertyNames(filtersTree).map((it) => (
          <div key={it}>
            {typeof filtersTree[it] === 'boolean' ? (
              <div>
                <Checkbox
                  size="small"
                  checked={isChecked(filtersTree[it])}
                  onChange={() => onToggle(it)}
                  className="pl-2"
                >
                  {it}
                </Checkbox>
              </div>
            ) : (
              <FoldablePanel caption={it} className="pb-3 pl-2">
                <FiltersTreeView
                  onToggle={onToggle}
                  filtersTree={filtersTree[it]}
                />
              </FoldablePanel>
            )}
          </div>
        ))}
    </div>
  )
}
