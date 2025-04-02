import { Table, HStack, Loader, ReadMore, Pagination } from '@navikt/ds-react'
import { CheckmarkCircleIcon, XMarkOctagonIcon } from '@navikt/aksel-icons';
import {
  WorkstationJobStateCompleted, WorkstationJobStateFailed,
  WorkstationZonalTagBindingsJob,
} from '../../lib/rest/generatedDto'
import humanizeDate from '../../lib/humanizeDate'
import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'

const ZonalTagBindingJobs = ({ jobs }: { jobs: WorkstationZonalTagBindingsJob[] }) => {
  const [page, setPage] = useState(1);
  const rowsPerPage = 4;


  let sortData = jobs;
  sortData = sortData.slice((page - 1) * rowsPerPage, page * rowsPerPage);

  return (
    <div className="flex flex-col gap-4 pt-4">
    <Table size="small">
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell scope="col">Jobb ID</Table.HeaderCell>
          <Table.HeaderCell scope="col">Status</Table.HeaderCell>
          <Table.HeaderCell scope="col">Tid for kjøring</Table.HeaderCell>
          <Table.HeaderCell scope="col"></Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {sortData.map((job) => (
          <Table.Row key={job.JobHeader.id}>
              <Table.DataCell>{job.JobHeader.id}</Table.DataCell>
              <Table.DataCell scope="row">
                {(() => {
                  switch (job.JobHeader.state) {
                    case WorkstationJobStateCompleted:
                      return (
                        <HStack gap="1">
                          <CheckmarkCircleIcon />
                          Completed
                        </HStack>
                      );
                    case WorkstationJobStateFailed:
                      return (
                        <HStack gap="1">
                          <XMarkOctagonIcon />
                          Failed
                        </HStack>
                      );
                    default:
                      return (
                        <HStack gap="1">
                          <Loader size="small" />
                          Running
                        </HStack>
                      );
                  }
                })()}
            </Table.DataCell>
            <Table.DataCell>{formatDistanceToNow(new Date(job.JobHeader.startTime), {addSuffix: true})}</Table.DataCell>
            <Table.DataCell>
              {job.JobHeader.errors.length > 0 && (
                  <ReadMore header="Feilmeldinger">
                    {job.JobHeader.errors.join(', ')}
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
      count={Math.ceil(jobs.length / rowsPerPage)}
      size="small"
    />
    </div>
  );
};

export default ZonalTagBindingJobs;
