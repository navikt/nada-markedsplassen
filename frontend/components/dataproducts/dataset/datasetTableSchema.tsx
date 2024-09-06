import * as React from 'react'
import { Alert, Heading, Table, Tag } from '@navikt/ds-react'
import { PIITagNames } from './useColumnTags'
import { PersonopplysningerDetaljert } from "./helptext";
import TagPill from '../../lib/tagPill';

interface DatasetTableSchemaProps {
  datasource: any
  showPii: boolean
}

const DatasetTableSchema = ({
  datasource,
  showPii,
}: DatasetTableSchemaProps) => {
  const schema = datasource.schema
  if (!schema) return <div>Ingen skjemainformasjon</div>

  return (
    <div>
      <Heading spacing level="3" size="small">
        Skjema
      </Heading>
      {!!datasource.pseudoColumns?.length && <Alert variant="info">Dette er et pseudoynimisert view</Alert>}
      <div className="mb-3 w-[91vw] overflow-auto">
        <Table className="w-fit 2xl:w-[60rem]" size="small">
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>Name</Table.HeaderCell>
              <Table.HeaderCell>Mode</Table.HeaderCell>
              <Table.HeaderCell>Type</Table.HeaderCell>
              <Table.HeaderCell>Description</Table.HeaderCell>
              {showPii && <Table.HeaderCell><span className="flex gap-2 items-center">Personopplysning <PersonopplysningerDetaljert /></span></Table.HeaderCell>}
              <Table.HeaderCell></Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {schema.map((row: any) => (
              <Table.Row key={row.name}>
                <Table.DataCell>{row.name}</Table.DataCell>
                <Table.DataCell>{row.mode}</Table.DataCell>
                <Table.DataCell>{row.type}</Table.DataCell>
                <Table.DataCell>{row.description}</Table.DataCell>
                {showPii && (
                  <Table.DataCell>
                    {PIITagNames.get(
                      JSON.parse(datasource.piiTags || '{}')[row.name]
                    )}
                  </Table.DataCell>
                )}
                <Table.DataCell>{datasource.pseudoColumns?.some((it: any) => it == row.name) && <Tag variant="warning">SHA256 hashed</Tag>}</Table.DataCell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </div>
    </div>
  )
}
export default DatasetTableSchema
