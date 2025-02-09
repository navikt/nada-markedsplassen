import {
  CheckmarkCircleIcon,
} from '@navikt/aksel-icons'
import { HStack, Heading, Loader, Table, Button } from '@navikt/ds-react'
import {
  WorkstationJobStateRunning,
  Workstation_STATE_RUNNING, WorkstationZonalTagBindingsJob,
} from '../../lib/rest/generatedDto'
import {
  useCreateZonalTagBindingsJob,
  useWorkstationEffectiveTags,
  useWorkstationMine, useWorkstationOnpremMapping,
  useWorkstationZonalTagBindingsJobs,
} from './queries'
import ZonalTagBindingJobs from './ZonalTagBindingJobs'

const WorkstationZonalTagBindings = ({}) => {
  const workstation = useWorkstationMine()
  const effectiveTags = useWorkstationEffectiveTags()
  const onpremMapping = useWorkstationOnpremMapping()
  const bindingJobs = useWorkstationZonalTagBindingsJobs()
  const createZonalTagBindingsJob = useCreateZonalTagBindingsJob();

  const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING
const handleCreateZonalTagBindingsJob = () => {
    if (onpremMapping.data) {
      createZonalTagBindingsJob.mutate(onpremMapping.data);
    }
  };
  if (!workstationIsRunning) {
    return <div></div>
  }

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
          <CheckmarkCircleIcon />
        </Table.DataCell>
      </Table.Row>
    )
  }

  return (
    <>
      <Heading className="pt-8" level="2" size="medium">Nettverk status</Heading>
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
      <Button disabled={bindingJobs.data?.jobs?.pop()?.state === WorkstationJobStateRunning || true} variant="primary" onClick={handleCreateZonalTagBindingsJob}>Koble til nettverk</Button>
      <Heading className="pt-8" level="2" size="medium">Oppkoblingsjobber</Heading>
      <ZonalTagBindingJobs jobs={bindingJobs.data?.jobs?.filter((job): job is WorkstationZonalTagBindingsJob => job !== undefined) || []} />
    </>
  )
}

export default WorkstationZonalTagBindings
