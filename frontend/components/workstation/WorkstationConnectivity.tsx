import {
  CheckmarkCircleIcon, CircleSlashIcon,
} from '@navikt/aksel-icons'
import { Alert, Button, Heading, HStack, Table } from '@navikt/ds-react'
import {
  JobStateRunning,
  Workstation_STATE_RUNNING, WorkstationConnectJob,
} from '../../lib/rest/generatedDto'
import ConnectivityWorkflow from './ConnectivityWorkflow'
import {
  useCreateWorkstationConnectivityWorkflow,
  useWorkstationConnectivityWorkflow,
  useWorkstationEffectiveTags,
  useWorkstationMine, useWorkstationOnpremMapping,
} from './queries'

import AnimatedCrate from './AnimatedCrate'
import Quiz from './Quiz'

const WorkstationConnectivity = ({}) => {
  const workstation = useWorkstationMine()
  const effectiveTags = useWorkstationEffectiveTags()
  const onpremMapping = useWorkstationOnpremMapping()
  const connectivityWorkflow = useWorkstationConnectivityWorkflow()
  const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow()

  const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING

  const handleCreateZonalTagBindingsJob = () => {
      if (onpremMapping.data) {
        createConnectivityWorkflow.mutate(onpremMapping.data);
      }
  };

  const handleDeleteZonalTagBindingsJob = () => {
    createConnectivityWorkflow.mutate({"hosts": []});
  };

  const hasRunningConnectJob: boolean = (connectivityWorkflow.data?.connect?.filter((job): job is WorkstationConnectJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;
  const hasRunningNotifyJob: boolean =  connectivityWorkflow.data?.notify?.state === JobStateRunning

  const hasRunningJob: boolean = hasRunningConnectJob || hasRunningNotifyJob
  const allSelectedInternalServicesAreActivated: boolean = effectiveTags.data?.tags?.length === onpremMapping.data?.hosts?.length;

  const renderStatus = (tag: string) => {
    const isEffective = effectiveTags.data?.tags?.some(eTag => eTag?.namespacedTagValue?.split('/').pop() === tag)

    if (isEffective) {
      return (
        <Table.Row key={tag}>
          <Table.HeaderCell scope="row">
            {tag}
          </Table.HeaderCell>
          <Table.DataCell>
            <HStack gap="1">
              Aktiv
            </HStack>
          </Table.DataCell>
          <Table.DataCell>
            <CheckmarkCircleIcon />
          </Table.DataCell>
        </Table.Row>
      )
    }

    return (
      <Table.Row key={tag}>
        <Table.HeaderCell scope="row">
          {tag}
        </Table.HeaderCell>
        <Table.DataCell>
          <HStack gap="1">
            Inaktiv
          </HStack>
        </Table.DataCell>
        <Table.DataCell>
          <CircleSlashIcon />
        </Table.DataCell>
      </Table.Row>
    )
  }

  return (
    <>
      <div className="flex flex-col gap-4 p-2 max-w-2xl">
        <AnimatedCrate><Quiz /></AnimatedCrate>
        <Alert variant="info">
          Dine valgte tjenester må aktiveres <b>hver gang du starter maskinen.</b>
        </Alert>
        </div>
      {(onpremMapping && onpremMapping.data && onpremMapping.data.hosts.length > 0) ? (
        <Table size="small">
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell scope="col">Åpning</Table.HeaderCell>
              <Table.HeaderCell scope="col">Status</Table.HeaderCell>
              <Table.HeaderCell scope="col"></Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {onpremMapping.data?.hosts.map(renderStatus)}
          </Table.Body>
        </Table>
      ) : (
        <p>Du har ikke bedt om noen nettverksåpninger.</p>
      )}

      {(connectivityWorkflow.data?.notify?.errors.length || 0) > 0 ? (
        <Alert variant="error" className="mb-4">
          Det oppstod en feil ved annonsering til datavarehuset: {connectivityWorkflow.data?.notify?.errors}
        </Alert>
      ) : null}
      <div className="flex flex-row gap-4 mt-4 w-auto">
        <Button
          disabled={hasRunningJob || !workstationIsRunning || allSelectedInternalServicesAreActivated}
          variant="primary"
          onClick={handleCreateZonalTagBindingsJob}
        >
          Aktiver valgte koblinger
        </Button>
        <Button
          disabled={hasRunningJob || !workstationIsRunning || !allSelectedInternalServicesAreActivated}
          variant="secondary"
          onClick={handleDeleteZonalTagBindingsJob}
        >
          Deaktiver valgte koblinger
        </Button>
      </div>
      <Heading className="pt-8" level="3" size="medium">Jobber</Heading>
      <ConnectivityWorkflow wf={connectivityWorkflow.data}/>
    </>
  )
}
export default WorkstationConnectivity
