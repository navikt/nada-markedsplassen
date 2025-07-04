import { Table, HStack, Loader, ReadMore, Pagination } from '@navikt/ds-react'
import { CheckmarkCircleIcon, XMarkOctagonIcon } from '@navikt/aksel-icons';
import {
  JobStateCompleted, JobStateFailed,
  WorkstationJob,
} from '../../lib/rest/generatedDto'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'

const WorkstationChanges = ({ jobs }: { jobs: WorkstationJob[] }) => {
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
            <Table.HeaderCell scope="col">Maskintype</Table.HeaderCell>
            <Table.HeaderCell scope="col">Utviklingsmiljø</Table.HeaderCell>
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
                          case JobStateCompleted:
                              return (
                                  <HStack gap="1">
                            <CheckmarkCircleIcon />
                            Completed
                          </HStack>
                        );
                        case JobStateFailed:
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
              <Table.DataCell>{job.machineType}</Table.DataCell>
              <Table.DataCell>{job.containerImage}</Table.DataCell>
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
        count={Math.ceil(jobs.length / rowsPerPage)}
        size="small"
      />
      </div>
    );
  };
  
  export default WorkstationChanges;
