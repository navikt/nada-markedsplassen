import { ChevronDownIcon, ChevronUpIcon, ExternalLinkFillIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import { Table } from "@navikt/ds-react";
import Link from "next/link";
import React from "react";

type KnastInfo = {
  operationalStatus: string;
  logs: string[];
}

export const InfoForm = ({ operationalStatus, logs }: KnastInfo) => {
  const [showDataSources, setShowDataSources] = React.useState(true);
  const [showInternetAccess, setShowInternetAccess] = React.useState(true);
  const [showAllLogs, setShowAllLogs] = React.useState(false);

  console.log("logs", logs)
  return <div className="max-w-[35rem] border-blue-100 border rounded p-4">
    <Table>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell colSpan={2} scope="col">Knast - Z123456</Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        <Table.Row>
          <Table.HeaderCell scope="row">Status</Table.HeaderCell>
          <Table.DataCell>
            <div className="flex flex-rol">
              {operationalStatus}
              {operationalStatus === "Running" && <Link href="#" className="flex flex-rol ml-2">Open<ExternalLinkIcon /></Link>}
            </div>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Local Dev (SSH)</Table.HeaderCell>
          <Table.DataCell className="flex flex-rol">Enabled <Link href="#" className="flex flex-rol ml-2">Read me!</Link></Table.DataCell>
        </Table.Row>

        <Table.Row>
          <Table.HeaderCell scope="row">Environment</Table.HeaderCell>
          <Table.DataCell>VS Code</Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Machine Type</Table.HeaderCell>
          <Table.DataCell>n2d-standard-2</Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Cost</Table.HeaderCell>
          <Table.DataCell>20 Kr/m</Table.DataCell>
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
            {logs?.slice(0, showAllLogs ? logs.length : 1).map((log, idx) => {
              return idx !== 0 ? <div key={idx}>{log}</div> : <div className="flex flex-rol" key={idx}>{log} {logs?.length > 1 && !showAllLogs ? <Link href="#" className="flex flex-rol ml-4" onClick={() => setShowAllLogs(true)}>More</Link> :
                logs?.length > 1 && showAllLogs ? <Link href="#" className="flex flex-rol ml-4" onClick={() => setShowAllLogs(false)}>Less</Link> : null
              }</div>

            })}
          </Table.DataCell>
        </Table.Row>

      </Table.Body>
    </Table>
  </div>
}