import * as React from 'react'
import { Select, Table, Checkbox, Popover, Alert } from '@navikt/ds-react'
import LoaderSpinner from '../../lib/spinner'
import {
  DEFAULT_COLUMN_TAG,
  PIITagOptions,
  PIITagType,
} from './useColumnTags'
import { PersonopplysningerDetaljert } from "./helptext";
import { useRef, useState } from 'react'
import { BigqueryColumn } from '../../../lib/rest/generatedDto';

interface PseudoCheckProps {
  pseudoColumns: Map<string, boolean>
  selectPseudoColumn?: (columnName: string, on: boolean) => void
  name: string
}

const PseudoCheck = ({ pseudoColumns, name, selectPseudoColumn }: PseudoCheckProps) => {
  const elRef = useRef<HTMLInputElement>(null);
  const [showAtLeastOne, setShowAtLeastOne] = useState(false)
  return <>
    <Checkbox checked={pseudoColumns.get(name)} ref={elRef} onChange={evt => {
      if (!Array.from(pseudoColumns).filter((e) => e[0] != name && e[1]).length && !evt.target.checked) {
        setShowAtLeastOne(true)
        return
      }
      setShowAtLeastOne(false)
      selectPseudoColumn?.(name, evt.target.checked)
    }}>{""}</Checkbox>      <Popover
      open={showAtLeastOne}
      placement="right"
      onClose={() => setShowAtLeastOne(false)}
      anchorEl={elRef.current}
    >
      <Popover.Content>Må ha minst en pseudokolonne</Popover.Content>
    </Popover></>
}


interface AnnotateDatasetTableProps {
  loading: boolean
  error: any
  columns: BigqueryColumn[] | undefined
  tags: Map<string, PIITagType> | undefined
  pseudoColumns: Map<string, boolean>
  annotateColumn: (columnName: string, tag: PIITagType) => void
  selectPseudoColumn?: (columnName: string, on: boolean) => void
}


const AnnotateDatasetTable = ({
  loading,
  error,
  columns,
  tags,
  pseudoColumns,
  selectPseudoColumn,
  annotateColumn,
}: AnnotateDatasetTableProps) => {

  if (loading) {
    return <LoaderSpinner />
  }

  if (error) {
    console.log(error)
    return <div>Kan ikke hent skjemainformasjon</div>
  }

  if (!columns) return <div>Ingen skjemainformasjon</div>

  const selectedAllColumns = Array.from(pseudoColumns).filter(e=> e[1]).length === columns.length
  return (
    <div className="mb-3 w-[91vw] overflow-auto">
      <p className="flex gap-2 items-center mb-2">Klassifiser personopplysningene <PersonopplysningerDetaljert /></p>
      <Table className="w-[60rem]" size="small">
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Mode</Table.HeaderCell>
            <Table.HeaderCell>Type</Table.HeaderCell>
            <Table.HeaderCell>Description</Table.HeaderCell>
            <Table.HeaderCell>Personopplysning</Table.HeaderCell>
            {selectPseudoColumn && <Table.HeaderCell>Pseudonymisèr</Table.HeaderCell>}

          </Table.Row>
        </Table.Header>
        <Table.Body>
          {columns?.map((row, index) => (
            <Table.Row key={row.name}>
              <Table.DataCell>{row.name}</Table.DataCell>
              <Table.DataCell>{row.mode}</Table.DataCell>
              <Table.DataCell>{row.type}</Table.DataCell>
              <Table.DataCell>{row.description}</Table.DataCell>
              <Table.DataCell className="w-60">
                <Select
                  className="w-full"
                  size="small"
                  label=""
                  value={
                    tags && tags.has(row.name)
                      ? tags.get(row.name)
                      : DEFAULT_COLUMN_TAG
                  }
                  onChange={(e) =>
                    annotateColumn(row.name, e.target.value as PIITagType)
                  }
                >
                  {Array.from(PIITagOptions).map(([key, name]) => (
                    <option value={key} key={key}>
                      {name}
                    </option>
                  ))}
                </Select>
              </Table.DataCell>
              {selectPseudoColumn && <Table.DataCell><PseudoCheck pseudoColumns={pseudoColumns} name={row.name} selectPseudoColumn={selectPseudoColumn}></PseudoCheck></Table.DataCell>}
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
      {selectedAllColumns && <Alert className="w-[60rem]" variant='error'>Skal ikke pseudonymisere alle kolonner</Alert>}
    </div>
  )
}
export default AnnotateDatasetTable
