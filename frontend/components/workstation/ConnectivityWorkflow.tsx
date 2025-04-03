import { Table, HStack, Loader, ReadMore, Pagination } from '@navikt/ds-react'
import { CheckmarkCircleIcon, XMarkOctagonIcon } from '@navikt/aksel-icons';
import {
  WorkstationConnectivityWorkflow, WorkstationConnectJob,
  WorkstationJobStateCompleted, WorkstationJobStateFailed,
} from '../../lib/rest/generatedDto'
import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'

type JobDetails = {
  id: number;
  type: string;
  state: string;
  details: string;
  startTime: string;
  errors: string[];
};

const ConnectivityWorkflow = ({ wf }: { wf: WorkstationConnectivityWorkflow | undefined}) => {
  const [page, setPage] = useState(1);
  const rowsPerPage = 5;

  let sortData: JobDetails[] = [
    ...(wf?.connect?.filter((job): job is WorkstationConnectJob => job !== undefined) || []).map((job) => ({
      id: job.id,
      type: 'Connect',
      state: job.state,
      details: job.host,
      startTime: job.startTime,
      errors: job.errors || []
    } as JobDetails)),
    ...(wf?.notify ? [{
      id: wf.notify.id,
      type: 'Notify',
      state: wf.notify.state,
      details: "Sjekk om DVH trenger beskjed om oppkobling.",
      startTime: wf.notify.startTime,
      errors: wf.notify.errors || []
    }] : []),
   ...(wf?.disconnect ?  [{
     id: wf.disconnect.id,
     type: 'Disconnect',
     state: wf.disconnect.state,
     details: "Koble fra tjenester som ikke er i bruk.",
     startTime: wf.disconnect.startTime,
     errors: wf.disconnect.errors || []
   }] : []),
  ];

  sortData = sortData.sort((a, b) => b.id - a.id)
  console.log('sortData', sortData)

  sortData = sortData.slice((page - 1) * rowsPerPage, page * rowsPerPage);

  return (
    <div className="flex flex-col gap-4 pt-4">
    <Table size="small">
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell scope="col">ID</Table.HeaderCell>
          <Table.HeaderCell scope="col">Type</Table.HeaderCell>
          <Table.HeaderCell scope="col">Status</Table.HeaderCell>
          <Table.HeaderCell scope="col">Detaljer</Table.HeaderCell>
          <Table.HeaderCell scope="col">Tid for kjøring</Table.HeaderCell>
          <Table.HeaderCell scope="col"></Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {sortData.map((job: JobDetails) => (
          <Table.Row key={job.id}>
              <Table.DataCell>{job.id}</Table.DataCell>
              <Table.DataCell>{job.type}</Table.DataCell>
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
            <Table.DataCell>{job.details}</Table.DataCell>
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
