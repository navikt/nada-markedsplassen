import {
  CheckmarkCircleIcon, CircleSlashIcon,
} from '@navikt/aksel-icons'
import { HStack, Heading, Table, Button, Alert, Modal, Link } from '@navikt/ds-react'
import {
  JobStateRunning,
  OnpremHostTypeTNS,
  Workstation_STATE_RUNNING, WorkstationConnectJob,
} from '../../lib/rest/generatedDto'
import ConnectivityWorkflow from './ConnectivityWorkflow'
import {
  useCreateWorkstationConnectivityWorkflow,
  useWorkstationConnectivityWorkflow,
  useWorkstationEffectiveTags,
  useWorkstationMine, useWorkstationOnpremMapping,
} from './queries'
import React from 'react'
import { useOnpremMapping } from '../onpremmapping/queries'
import { configWorkstationSSH } from '../../lib/rest/workstation'

import AnimatedCrate from './AnimatedCrate'
import Quiz from './Quiz'

const WorkstationConnectivity = ({}) => {
  const workstation = useWorkstationMine()
  const effectiveTags = useWorkstationEffectiveTags()
  const workstationOnpremMapping = useWorkstationOnpremMapping()
  const onpremMapping = useOnpremMapping()
  const connectivityWorkflow = useWorkstationConnectivityWorkflow()
  const createConnectivityWorkflow = useCreateWorkstationConnectivityWorkflow()

  const workstationIsRunning = workstation.data?.state === Workstation_STATE_RUNNING
  const [openDVHAlert, setOpenDVHAlert] = React.useState(false);

  const handleCreateZonalTagBindingsJob = () => {
      if (workstation.data?.allowSSH && workstationOnpremMapping.data?.hosts?.some(
        h => Object.entries(onpremMapping.data?.hosts??{}).find(([type, _]) => type === OnpremHostTypeTNS)?.[1].some(host => host?.Host === h))) {
        setOpenDVHAlert(true);
        return;
      }
      if (workstationOnpremMapping.data) {
        createConnectivityWorkflow.mutate(workstationOnpremMapping.data);
      }
  };

  const handleDeleteZonalTagBindingsJob = () => {
    createConnectivityWorkflow.mutate({"hosts": []});
  };

  const hasRunningConnectJob: boolean = (connectivityWorkflow.data?.connect?.filter((job): job is WorkstationConnectJob => job !== undefined && job.state === JobStateRunning).length || 0) > 0;
  const hasRunningNotifyJob: boolean =  connectivityWorkflow.data?.notify?.state === JobStateRunning

  const hasRunningJob: boolean = hasRunningConnectJob || hasRunningNotifyJob
  const allSelectedInternalServicesAreActivated: boolean = effectiveTags.data?.tags?.length === workstationOnpremMapping.data?.hosts?.length;

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
    <Modal open={openDVHAlert} onClose={() => setOpenDVHAlert(false)}
      aria-label="">
      <Modal.Body>
        Av sikkerhetshensyn kan ikke Knast åpne DVH-kilder når SSH (lokal IDE-tilgang) er aktivert. Du kan enten  <Link href="#" onClick={() =>{
                          configWorkstationSSH(false)
                        setOpenDVHAlert(false)
        }
                        }>deaktivere SSH</Link> eller fjerne DVH-kilder.
      </Modal.Body>
      <Modal.Footer>
          <Button type="button" onClick={() => setOpenDVHAlert(false)}>
            Lukk
          </Button>
      </Modal.Footer>
    </Modal>
      <div className="flex flex-col gap-4 p-2 max-w-2xl">
        <AnimatedCrate><Quiz /></AnimatedCrate>
        <Alert variant="info">
          Dine valgte tjenester må aktiveres <b>hver gang du starter maskinen.</b>
        </Alert>
        </div>
      {(workstationOnpremMapping && workstationOnpremMapping.data && workstationOnpremMapping.data.hosts.length > 0) ? (
        <Table size="small">
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell scope="col">Åpning</Table.HeaderCell>
              <Table.HeaderCell scope="col">Status</Table.HeaderCell>
              <Table.HeaderCell scope="col"></Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {workstationOnpremMapping.data?.hosts.map(renderStatus)}
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
