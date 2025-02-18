import { Button, Checkbox, CheckboxGroup, Heading, UNSAFE_Combobox } from '@navikt/ds-react'
import { useImperativeHandle, useRef, useState } from 'react'
import {
  useCreateZonalTagBindingsJob,
  useUpdateWorkstationOnpremMapping,
  useWorkstationOnpremMapping,
} from '../queries'
import { useOnpremMapping } from '../../onpremmapping/queries'
import {
  Host,
  OnpremHostTypeHTTP, OnpremHostTypeInformatica, OnpremHostTypeOracle,
  OnpremHostTypePostgres, OnpremHostTypeSFTP, OnpremHostTypeSMTP,
  OnpremHostTypeTNS,
} from '../../../lib/rest/generatedDto'

interface FirewallTagSelectorProps {
  enabled?: boolean;
}

interface HostsProps {
  enabled?: boolean;
  title: string;
  hosts: Host[];
  preselected: string[];
  ref: React.Ref<{ getSelectedHosts: () => string[] }>;
}

const HostsList = (props: HostsProps) => {
  const [selectedHosts, setSelectedHosts] = useState<Set<string>>(new Set(props.preselected))

  const handleChange = (host: string, isSelected: boolean) => {
    if (isSelected) {
      setSelectedHosts(new Set(selectedHosts.add(host)))
    } else {
      const newSelectedHosts = new Set(selectedHosts);
      newSelectedHosts.delete(host);
      setSelectedHosts(newSelectedHosts);
    }
  }

  useImperativeHandle(props.ref, () => ({
    getSelectedHosts: () => {
     return Array.from(selectedHosts)
    },
  }))

  return (
    <UNSAFE_Combobox
      isMultiSelect
      hideLabel
      readOnly={!props.enabled}
      label={props.title}
      options={props.hosts.map((host) => ({ label: host.Name, value: host.Host }))}
      selectedOptions={Array.from(selectedHosts)}
      onToggleSelected={handleChange}
    />
  )
}

const HostsChecked = (props: HostsProps) => {
  const [selectedHosts, setSelectedHosts] = useState<string[]>(props.preselected)

  useImperativeHandle(props.ref, () => ({
    getSelectedHosts: () => {
     return selectedHosts
    },
  }))

  return (
    <div>
      <CheckboxGroup disabled={!props.enabled} legend={props.title} hideLegend size="small"  onChange={setSelectedHosts} value={selectedHosts}>
        {props.hosts.map((host) => (
          <Checkbox description={host.Description} key={host.Name} value={host.Host}>{host.Name}</Checkbox>
        ))}
      </CheckboxGroup>
    </div>
  )
}

export const FirewallTagSelector = (props: FirewallTagSelectorProps) => {
  const onpremMapping = useOnpremMapping()
  const workstationOnpremMapping = useWorkstationOnpremMapping()

  const updateWorkstationOnpremMapping = useUpdateWorkstationOnpremMapping()

  const preselectedHosts = workstationOnpremMapping.data?.hosts || []
  const onpremHosts = onpremMapping.data?.hosts

  const postgresRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const tnsRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const httpRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const sftpRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const informaticaRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const oracleRef = useRef<{getSelectedHosts: () => string[] }>(null)
  const smtpRef = useRef<{getSelectedHosts: () => string[] }>(null)

  const submit = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault()

    const selectedPostgresHost = postgresRef.current?.getSelectedHosts?.()
    const selectedTnsHosts = tnsRef.current?.getSelectedHosts?.()
    const selectedHttpHosts = httpRef.current?.getSelectedHosts?.()
    const selectedSftpHosts = sftpRef.current?.getSelectedHosts?.()
    const selectedInformaticaHosts = informaticaRef.current?.getSelectedHosts?.()
    const selectedOracleHosts = oracleRef.current?.getSelectedHosts?.()
    const selectedSmtpHosts = smtpRef.current?.getSelectedHosts?.()

    const uniqueHosts = Array.from(new Set([
      ...selectedPostgresHost || [],
      ...selectedTnsHosts || [],
      ...selectedHttpHosts || [],
      ...selectedSftpHosts || [],
      ...selectedInformaticaHosts || [],
      ...selectedOracleHosts || [],
      ...selectedSmtpHosts || [],
    ]))

    try {
      updateWorkstationOnpremMapping.mutate({
        hosts: uniqueHosts,
      })
    } catch (error) {
      console.error('Failed to update onprem mapping:', error)
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <div>
        Styr nettverksforbindelser mot on-prem tjenester du ønsker å koble opp mot.
        Du må starte maskinen for å kunne koble opp mot on-prem tjenester.
      </div>
      {onpremHosts && Object.entries(onpremHosts).sort(([typeA], [typeB]) => {
        // Define your custom sorting logic here
        const order = [
          OnpremHostTypeTNS,
          OnpremHostTypePostgres,
          OnpremHostTypeHTTP,
          OnpremHostTypeSFTP,
          OnpremHostTypeInformatica,
          OnpremHostTypeOracle,
          OnpremHostTypeSMTP,
        ];
        return order.indexOf(typeA) - order.indexOf(typeB);
      }).map(([type, hosts]) => {
        const preselected = preselectedHosts.filter(host => hosts.some(h => h != undefined && h.Host === host))
        switch (type) {
          case OnpremHostTypeTNS:
            return (
              <div key={type}>
                <Heading size="xsmall">Datavarehus</Heading>
                <HostsChecked enabled={props.enabled} title="Datavarehus" ref={tnsRef} preselected={preselected}
                              hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypePostgres:
            return (
              <div key={type}>
                <Heading size="xsmall">Postgres</Heading>
                <HostsList enabled={props.enabled} title="Postgres" ref={postgresRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypeHTTP:
            return (
              <div key={type}>
                <Heading size="xsmall">HTTPS</Heading>
                <HostsChecked enabled={props.enabled} title="HTTPS" ref={httpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypeSFTP:
            return (
              <div key={type}>
                <Heading size="xsmall">SFTP</Heading>
                <HostsList enabled={props.enabled} title="SFTP" ref={sftpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypeInformatica:
            return (
              <div key={type}>
                <Heading size="xsmall">Informatica</Heading>
                <HostsList enabled={props.enabled} title="Informatica" ref={informaticaRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypeOracle:
            return (
              <div key={type}>
                <Heading size="xsmall">Oracle</Heading>
                <HostsList enabled={props.enabled} title="Oracle" ref={oracleRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          case OnpremHostTypeSMTP:
            return (
              <div key={type}>
                <Heading size="xsmall">SMTP</Heading>
                <HostsList enabled={props.enabled} title="SMTP" ref={smtpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} />
              </div>
            )
          default:
            return null
        }
      })}
      <Button disabled={!props.enabled} onClick={submit} variant="primary">Lagre</Button>
    </div>
  )
}


export default FirewallTagSelector
