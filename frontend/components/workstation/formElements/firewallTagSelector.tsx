import { Checkbox, CheckboxGroup, Heading, UNSAFE_Combobox } from '@navikt/ds-react'
import { useImperativeHandle, useRef, useState } from 'react'
import { useUpdateWorkstationOnpremMapping, useWorkstationOnpremMapping } from '../queries'
import { useOnpremMapping } from '../../onpremmapping/queries'
import {
  Host,
  OnpremHostTypeHTTP,
  OnpremHostTypeInformatica,
  OnpremHostTypeOracle,
  OnpremHostTypePostgres,
  OnpremHostTypeSFTP,
  OnpremHostTypeSMTP,
  OnpremHostTypeTNS,
  OnpremHostTypeCloudSQL,
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
  submit: () => void;
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

    props.submit()
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

  const handleChange = (hosts: string[]) => {
    setSelectedHosts(hosts)
    props.submit()
  }

  useImperativeHandle(props.ref, () => ({
    getSelectedHosts: () => {
      const checkboxValues = document.querySelectorAll('input[type="checkbox"]:checked');
      return Array.from(checkboxValues).map(el => (el as HTMLInputElement).value)
    },
  }))

  return (
    <div>
      <CheckboxGroup disabled={!props.enabled} legend={props.title} hideLegend size="small"  onChange={handleChange} value={selectedHosts}>
        {props.hosts.map((host) => (
          <Checkbox description={host.Description} key={host.Name} value={host.Host}>{host.Name} {host.Name !== host.Host ? "(" + host.Host + ")" : ""}</Checkbox>
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
  const cloudsqlRef = useRef<{getSelectedHosts: () => string[] }>(null)

  const submit = () => {
    const selectedPostgresHost = postgresRef.current?.getSelectedHosts?.()
    const selectedTnsHosts = tnsRef.current?.getSelectedHosts?.()
    const selectedHttpHosts = httpRef.current?.getSelectedHosts?.()
    const selectedSftpHosts = sftpRef.current?.getSelectedHosts?.()
    const selectedCloudSQLHosts = cloudsqlRef.current?.getSelectedHosts?.()
    const selectedInformaticaHosts = informaticaRef.current?.getSelectedHosts?.()
    const selectedOracleHosts = oracleRef.current?.getSelectedHosts?.()
    const selectedSmtpHosts = smtpRef.current?.getSelectedHosts?.()

    const uniqueHosts = Array.from(new Set([
      ...selectedPostgresHost || [],
      ...selectedTnsHosts || [],
      ...selectedHttpHosts || [],
      ...selectedSftpHosts || [],
      ...selectedCloudSQLHosts || [],
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
    <div className="flex flex-col gap-12">
      <div/>
      <Heading level="2" size="medium">On-prem tjenester</Heading>
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
          OnpremHostTypeCloudSQL,
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
                <Heading size="small">Datavarehus</Heading>
                <HostsChecked enabled={props.enabled} title="Datavarehus" ref={tnsRef} preselected={preselected}
                              hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypePostgres:
            return (
              <div key={type}>
                <Heading size="small">Postgres</Heading>
                <HostsList enabled={props.enabled} title="Postgres" ref={postgresRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypeHTTP:
            return (
              <div key={type}>
                <Heading size="small">HTTPS</Heading>
                <HostsChecked enabled={props.enabled} title="HTTPS" ref={httpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypeSFTP:
            return (
              <div key={type}>
                <Heading size="small">SFTP</Heading>
                <HostsList enabled={props.enabled} title="SFTP" ref={sftpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypeInformatica:
            return (
              <div key={type}>
                <Heading size="small">Informatica</Heading>
                <HostsList enabled={props.enabled} title="Informatica" ref={informaticaRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypeOracle:
            return (
              <div key={type}>
                <Heading size="small">Oracle</Heading>
                <HostsList enabled={props.enabled} title="Oracle" ref={oracleRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
          case OnpremHostTypeSMTP:
            return (
              <div key={type}>
                <Heading size="small">SMTP</Heading>
                <HostsList enabled={props.enabled} title="SMTP" ref={smtpRef} preselected={preselected}
                           hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
              </div>
            )
            case OnpremHostTypeCloudSQL:
                return (
                  <div key={type}>
                    <Heading size="small">CloudSQL</Heading>
                    <HostsList enabled={props.enabled} title="CloudSQL" ref={cloudsqlRef} preselected={preselected}
                               hosts={hosts.filter((host): host is Host => host !== undefined)} submit={submit}/>
                  </div>
                )
          default:
            return null
        }
      })}
    </div>
  )
}


export default FirewallTagSelector
