import { ChevronDownIcon, ChevronUpIcon, ExternalLinkFillIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import { Table } from "@navikt/ds-react";
import Link from "next/link";
import React from "react";
import { Workstation_STATE_RUNNING, WorkstationOutput } from "../../lib/rest/generatedDto";
import { GetKnastDailyCost, GetOperationalStatus } from "./utils";
import { OpenKnastLink } from "./widgets/openKnastLink";
import { ColorAuxText, ColorDisabled } from "./designTokens";
import { IconConnectLightGray, IconConnectLightGreen, IconConnectLightRed, IconGear } from "./widgets/knastIcons";

type KnastInfo = {
  knastInfo: any
  operationalStatus?: string
}

const operationStatusText = new Map<string, string>([
  ["Started", "Kjører"],
  ["Stopped", "Stoppet"],
  ["Starting", "Starter"],
  ["Stopping", "Stopper"],
])

export const InfoForm = ({ knastInfo, operationalStatus }: KnastInfo) => {
  const [showAllDataSources, setShowAllDataSources] = React.useState(false);
  const [showAllLogs, setShowAllLogs] = React.useState(false);

  //TODO: is it possible that the tags.length!=0 && tags.length != hosts.length?
  const onpremActivated = knastInfo.workstationOnpremMapping && knastInfo.workstationOnpremMapping.hosts && knastInfo.workstationOnpremMapping.hosts.length > 0
  && knastInfo.effectiveTags && knastInfo.effectiveTags.tags && knastInfo.effectiveTags.tags.length === knastInfo.workstationOnpremMapping.hosts.length;

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
          <Table.DataCell>
            <div>
            <div>
            {
              knastInfo.workstationOnpremMapping?knastInfo.workstationOnpremMapping.hosts?.length>0?
                knastInfo.workstationOnpremMapping.hosts.slice(0, showAllDataSources? knastInfo.workstationOnpremMapping.hosts.length: 3)
                  .map((mapping, index) => (
                    <div className="grid grid-cols-[20px_1fr] items-center">
                      {onpremActivated? <IconConnectLightGreen />: <IconConnectLightRed /> }
                      <div key={index}>{mapping}</div>
                    </div>
                ))
              :<div className="flex flex-row items-center space-x-2"><div>{"Ikke konfigurert"}</div><Link href="#"><IconGear width={20} height={20}/></Link></div>: undefined
            }
            </div>
            {knastInfo.workstationOnpremMapping && knastInfo.workstationOnpremMapping.hosts.length > 3 &&
              <button className="mt-2 text-sm text-blue-600 hover:underline" onClick={() => setShowAllDataSources(!showAllDataSources)}>
                {showAllDataSources ? "Vis færre" : `Vis alle (${knastInfo.workstationOnpremMapping.hosts.length})`}
              </button>
            }
            </div>
          </Table.DataCell>
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