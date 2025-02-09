import {
  CheckmarkCircleIcon,
} from '@navikt/aksel-icons'
import { HStack, Heading, Loader, Table } from '@navikt/ds-react'
import {
  WorkstationJobStateRunning,
  Workstation_STATE_RUNNING,
} from '../../lib/rest/generatedDto'
import {
  useWorkstationEffectiveTags,
  useWorkstationMine, useWorkstationOnpremMapping,
  useWorkstationZonalTagBindingsJobs,
} from './queries'

const WorkstationZonalTagBindings = ({}) => {
  const workstation = useWorkstationMine()
  const effectiveTags = useWorkstationEffectiveTags()
  const onpremMapping = useWorkstationOnpremMapping()
  const bindingJobs = useWorkstationZonalTagBindingsJobs()

  const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING

  if (!workstationIsRunning) {
    return <div></div>
  }

  const runningJobs = bindingJobs.data?.jobs?.filter(job => job != undefined && job.state === WorkstationJobStateRunning) ?? []

  if (bindingJobs.isLoading || runningJobs.length > 0) {
    const firstRunningJob = runningJobs[0];
    const jobErrors = firstRunningJob?.errors ?? [];

    return (
      <div>
        <Heading className="pt-8" level="2" size="medium">Nettverk status</Heading>
        <div>
          Kobler til nettverk... <Loader title="En oppkoblings jobb kjører i bakgrunnen!" size="small" />
        </div>
        {jobErrors.length > 0 ? (
          <div>
            <p>Feil under oppkobling :(</p>
            <ul>
              {jobErrors.map((error, index) => (
                <li key={index}>{error}</li>
              ))}
            </ul>
          </div>
        ) : null}
      </div>
    );
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
    </>
  )
}

export default WorkstationZonalTagBindings
