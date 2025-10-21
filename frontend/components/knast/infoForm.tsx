import { ChevronDownIcon, ChevronUpIcon, ExternalLinkFillIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import { Table } from "@navikt/ds-react";
import Link from "next/link";
import React from "react";
import { Workstation_STATE_RUNNING, WorkstationOutput } from "../../lib/rest/generatedDto";
import { GetKnastDailyCost, GetOperationalStatus } from "./utils";
import { OpenKnastLink } from "./widgets/openKnastLink";
import { ColorAuxText, ColorDisabled } from "./designTokens";

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
  const [showDataSources, setShowDataSources] = React.useState(true);
  const [showInternetAccess, setShowInternetAccess] = React.useState(true);
  const [showAllLogs, setShowAllLogs] = React.useState(false);

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
          <Table.HeaderCell scope="row">Last Used</Table.HeaderCell>
          <Table.DataCell>Sep 2, 2011, 11:04 </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Data Sources</Table.HeaderCell>
          <Table.DataCell>
            <div>
              {!showDataSources && <Link href="#" className="flex flex-rol" onClick={() => setShowDataSources(true)}>Show<ChevronDownIcon /></Link>}
              {showDataSources && <Link href="#" className="flex flex-rol" onClick={() => setShowDataSources(false)}>Hide<ChevronUpIcon /></Link>}
              {showDataSources && <ul className="list-disc list-inside mt-2">
                <li>dm08-scan.adeo.no</li>
                <li>dmv34-scan.adeo.no</li>
              </ul>}
            </div>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Internet Access</Table.HeaderCell>
          <Table.DataCell>
            <div>
              {!showInternetAccess && <Link href="#" className="flex flex-rol" onClick={() => setShowInternetAccess(true)}>Show<ChevronDownIcon /></Link>}
              {showInternetAccess && <Link href="#" className="flex flex-rol" onClick={() => setShowInternetAccess(false)}>Hide<ChevronUpIcon /></Link>}
              {showInternetAccess && <ul className="list-disc list-inside mt-2">
                <li>vg.no</li>
                <li>power.no</li>
                <li>yr.no</li>
              </ul>}
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