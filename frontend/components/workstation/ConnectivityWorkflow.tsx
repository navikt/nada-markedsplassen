import { Table, HStack, Loader, ReadMore, Pagination } from '@navikt/ds-react'
import { CheckmarkCircleIcon, XMarkOctagonIcon } from '@navikt/aksel-icons';
import {
  WorkstationConnectivityWorkflow, WorkstationConnectJob,
  WorkstationJobStateCompleted, WorkstationJobStateFailed,
} from '../../lib/rest/generatedDto'
import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'

const ConnectivityWorkflow = ({ wf }: { wf: WorkstationConnectivityWorkflow | undefined}) => {
  const [page, setPage] = useState(1);
  const rowsPerPage = 10;

  let sortData = wf?.connect.filter((job): job is WorkstationConnectJob => job !== undefined) || []
  sortData = sortData.slice((page - 1) * rowsPerPage, page * rowsPerPage);

  return (
    <div className="flex flex-col gap-4 pt-4">
    <Table size="small">
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell scope="col">Jobb ID</Table.HeaderCell>
          <Table.HeaderCell scope="col">Status</Table.HeaderCell>
          <Table.HeaderCell scope="col">Tjeneste</Table.HeaderCell>
          <Table.HeaderCell scope="col">Tid for kjøring</Table.HeaderCell>
          <Table.HeaderCell scope="col"></Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {sortData.map((job) => (
          <Table.Row key={job.id}>
              <Table.DataCell>{job.id}</Table.DataCell>
              <Table.DataCell scope="row">
                {(() => {
                  switch (job.state) {
                    case WorkstationJobStateCompleted:
                      return (
                        <HStack gap="1">
                          <CheckmarkCircleIcon />
                          Fullført
                        </HStack>
                      );
                    case WorkstationJobStateFailed:
                      return (
                        <HStack gap="1">
                          <XMarkOctagonIcon />
                          Feilet
                        </HStack>
                      );
                    default:
                      return (
                        <HStack gap="1">
                          <Loader size="small" />
                          Kjører
                        </HStack>
                      );
                  }
                })()}
            </Table.DataCell>
            <Table.DataCell>{job.host}</Table.DataCell>
            <Table.DataCell>{formatDistanceToNow(new Date(job.startTime), {addSuffix: true})}</Table.DataCell>
            <Table.DataCell>
              {job.errors.length > 0 && (
                  <ReadMore header="Feilmeldinger">
                    {job.errors.join(', ')}
                  </ReadMore>
              )}
            </Table.DataCell>
          </Table.Row>
        ))}
      </Table.Body>
    </Table>
    <Pagination
      page={page}
      onPageChange={setPage}
      count={Math.ceil((wf?.connect?.length || 0) / rowsPerPage)}
      size="small"
    />
    </div>
  );
};

export default ConnectivityWorkflow;
