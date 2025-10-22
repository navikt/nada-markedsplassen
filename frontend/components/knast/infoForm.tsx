import { ChevronDownIcon, ChevronUpIcon, ExternalLinkFillIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import { Loader, Table } from "@navikt/ds-react";
import Link from "next/link";
import React from "react";
import { Workstation_STATE_RUNNING, WorkstationOutput } from "../../lib/rest/generatedDto";
import { GetKnastDailyCost, GetOperationalStatus } from "./utils";
import { OpenKnastLink } from "./widgets/openKnastLink";
import { ColorAuxText, ColorDisabled } from "./designTokens";
import { IconConnectLightGray, IconConnectLightGreen, IconConnectLightRed, IconGear } from "./widgets/knastIcons";
import { kn } from "date-fns/locale";

type InfoFormProps = {
  knastInfo: any
  operationalStatus?: string
  onActivateOnprem: () => void
  onActivateInternet: () => void
  onDeactivateOnPrem: () => void
  onDeactivateInternet: () => void
  onConfigureOnprem: () => void
  onConfigureInternet: () => void
}

const operationStatusText = new Map<string, string>([
  ["started", "Kjører"],
  ["stopped", "Stoppet"],
  ["starting", "Starter"],
  ["stopping", "Stopper"],
])

export const InfoForm = ({ knastInfo, operationalStatus, onActivateOnprem, onActivateInternet, onDeactivateOnPrem, onDeactivateInternet, onConfigureOnprem, onConfigureInternet }: InfoFormProps) => {
  const [showAllDataSources, setShowAllDataSources] = React.useState(false);
  const [showAllLogs, setShowAllLogs] = React.useState(false);

  console.log("knastInfo in InfoForm", knastInfo);
  const OnpremList = () => (<div>
    {
      knastInfo.workstationOnpremMapping ? knastInfo.workstationOnpremMapping.hosts?.length > 0 ?
        knastInfo.workstationOnpremMapping.hosts.slice(0, showAllDataSources ? knastInfo.workstationOnpremMapping.hosts.length : 3)
          .map((mapping: any, index: number) => (
            <div className="grid grid-cols-[20px_1fr] items-center">
              {knastInfo.allowSSH || knastInfo.operationalStatus!== "started" ? <IconConnectLightGray /> 
              : knastInfo.effectiveTags?.tags?.find((tag: any)=> tag.namespacedTagKey?.split("/").pop()=== mapping) ? <IconConnectLightGreen /> : <IconConnectLightRed />}
              <div key={index}>{mapping}</div>
            </div>
          ))
        : <div>{"Ikke konfigurert"}</div> : undefined
    }
    <div className="flex flex-row space-x-4 mt-2">
      {knastInfo.workstationOnpremMapping && knastInfo.workstationOnpremMapping.hosts.length > 3 &&
        <button className="text-sm text-blue-600 hover:underline" onClick={() => setShowAllDataSources(!showAllDataSources)}>
          {showAllDataSources ? "Vis færre" : `Vis alle (${knastInfo.workstationOnpremMapping.hosts.length})`}
        </button>
      }
      <div className="flex flex-row items-center space-x-4">
        <button className="text-sm text-blue-600 hover:underline" onClick={() => { }}>
          Konfigurer
        </button>
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => knastInfo.onpremState === "activated" ? onDeactivateOnPrem() : knastInfo.onpremState === "deactivated" ? onActivateOnprem() : undefined}
          disabled={knastInfo.onpremState === "deactivating" || knastInfo.onpremState === "activating"}
          hidden={knastInfo.operationalStatus !== "started" || !knastInfo.onpremConfigured || !knastInfo.onpremState || knastInfo.allowSSH
            || knastInfo.onpremState === "activating" || knastInfo.onpremState === "deactivating"
          }>
          {knastInfo.onpremState === "activated" ? "Deaktiver" : knastInfo.onpremState === "deactivated" ? "Aktiver" : ""}
        </button>
        {(knastInfo.onpremState === "activating" || knastInfo.onpremState === "deactivating") && <div className="flex flex-row">
          <div className="text-sm" style={{ color: ColorAuxText }}>oppdater</div>
          <Loader size="small" />
        </div>}
      </div>
    </div>
  </div>);


  console.log(knastInfo)
  return <div className="max-w-[35rem] border-blue-100 border rounded p-4">
    <Table>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell colSpan={2} scope="col">Knast - {knastInfo.displayName}</Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        <Table.Row>
          <Table.HeaderCell scope="row">Status</Table.HeaderCell>
          <Table.DataCell>
            {operationStatusText.get(operationalStatus ?? "") || "Ukjent"}
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Miljø</Table.HeaderCell>
          <Table.DataCell>
            <div className="flex flex-rol">
              {knastInfo.imageTitle}
              {operationalStatus === "Started" && <div className="pl-4"><OpenKnastLink caption={"Åpne"} knastInfo={knastInfo} /></div>}
            </div>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Maskintype</Table.HeaderCell>
          <Table.DataCell>
            <div>{knastInfo.machineTypeInfo?.machineType || "Ukjent"}</div>
            <div className="text-sm" style={{
              color: ColorAuxText
            }}>{knastInfo.machineTypeInfo && `${knastInfo.machineTypeInfo.vCPU} vCPU, ${knastInfo.machineTypeInfo.memoryGB} GB RAM`}</div>
          </Table.DataCell>
        </Table.Row>

        <Table.Row>
          <Table.HeaderCell scope="row">Lokal dev (SSH)</Table.HeaderCell>
          <Table.DataCell className="flex flex-rol">{knastInfo.allowSSH ? "Aktivert" : "Deaktivert"}
            {knastInfo.allowSSH && <Link href="#" className="flex flex-rol ml-2">Guide</Link>
            }</Table.DataCell>
        </Table.Row>

        <Table.Row>
          <Table.HeaderCell scope="row">Kostnad</Table.HeaderCell>
          <Table.DataCell>{GetKnastDailyCost(knastInfo) || "Ukjent"}</Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Nav datakilder</Table.HeaderCell>
          <Table.DataCell> <OnpremList /> </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Internet Access</Table.HeaderCell>
          <Table.DataCell>
            <div>
              <ul className="list-disc list-inside mt-2">
                <li>vg.no</li>
                <li>power.no</li>
                <li>yr.no</li>
              </ul>
            </div>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Events
          </Table.HeaderCell>
          <Table.DataCell>
          </Table.DataCell>
        </Table.Row>

      </Table.Body>
    </Table>
  </div>
}