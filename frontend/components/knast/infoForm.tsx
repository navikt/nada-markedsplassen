import { ChevronDownIcon, ChevronUpIcon, CircleSlashIcon, ExclamationmarkTriangleIcon, ExternalLinkFillIcon, ExternalLinkIcon, InformationIcon, InformationSquareFillIcon } from "@navikt/aksel-icons";
import { Checkbox, Loader, Table, Tooltip } from "@navikt/ds-react";
import Link from "next/link";
import React from "react";
import { Workstation_STATE_RUNNING, WorkstationOutput } from "../../lib/rest/generatedDto";
import { getKnastDailyCost, getOperationalStatus } from "./utils";
import { OpenKnastLink } from "./widgets/openKnastLink";
import { ColorAuxText, ColorDefaultText, ColorDisabled, ColorFailed, ColorInfo, ColorSuccessful } from "./designTokens";
import { IconConnected, IconConnectLightGray, IconConnectLightGreen, IconConnectLightRed, IconDisconnected, IconGear } from "./widgets/knastIcons";
import { kn } from "date-fns/locale";
import { useUpdateWorkstationURLListItemForIdent } from "./queries";
import { formatDate } from "date-fns";
import { all } from "deepmerge";
import { LocalDevInfo } from "./widgets/localdevInfo";

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
  const updateUrlItem = useUpdateWorkstationURLListItemForIdent();
  const [showLocalDevInfo, setShowLocalDevInfo] = React.useState(false);
  const allOnpremHosts = [
    ...knastInfo?.workstationOnpremMapping,
    ...knastInfo?.effectiveTags?.tags?.filter((it: any) => !knastInfo?.workstationOnpremMapping?.some((mapping: any) => mapping.host === it.namespacedTagKey?.split("/").pop())).map((it: any) => {
      return { host: it.namespacedTagKey?.split("/").pop(), isDVHSource: false }
    }) || []
  ]

  const OnpremList = () => (<div>
    {
      allOnpremHosts.length > 0 ?
        allOnpremHosts.slice(0, showAllDataSources ? allOnpremHosts.length : 5)
          .map((mapping: any, index: number) => (
            <div key={index} className="grid grid-cols-[20px_1fr] items-center">
              {knastInfo.operationalStatus !== "started" ? <IconConnectLightGray />
                : knastInfo.effectiveTags?.tags?.find((tag: any) => tag.namespacedTagKey?.split("/").pop() === mapping.host)
                  ? <IconConnected width={12} />
                  : mapping.isDVHSource && knastInfo.allowSSH ? <IconConnectLightGray /> : <IconDisconnected width={12} />}
              <Tooltip hidden={(!mapping.isDVHSource || !knastInfo.allowSSH) && (operationalStatus === "started")} 
              content={operationalStatus === "started" ? "Denne kilden er en DVH-kilde og kan ikke nås når SSH er aktivert" : "Du kan ikke aktivere tilkoblinger når knast ikke er startet"}>
                <div key={index} style={{
                  color: mapping.isDVHSource && knastInfo.allowSSH ? ColorDisabled : ColorDefaultText
                }}>{mapping.host}</div>
              </Tooltip>
            </div>
          ))
        : <div>{"Ikke konfigurert"}</div>
    }
    <div className="flex flex-row space-x-4 mt-2">
      {knastInfo.workstationOnpremMapping && knastInfo.workstationOnpremMapping.length > 5 &&
        <button className="text-sm text-blue-600 hover:underline" onClick={() => setShowAllDataSources(!showAllDataSources)}>
          {showAllDataSources ? "Vis færre" : `Vis alle (${knastInfo.workstationOnpremMapping.length})`}
        </button>
      }
      <div className="flex flex-row items-center space-x-4">
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => knastInfo.onpremState === "activated" ? onDeactivateOnPrem() : knastInfo.onpremState === "deactivated" ? onActivateOnprem() : undefined}
          disabled={knastInfo.onpremState === "updating"}
          hidden={knastInfo.operationalStatus !== "started" || !allOnpremHosts.length || !knastInfo.onpremState
            || knastInfo.onpremState === "updating"
          }>
          {knastInfo.onpremState === "activated" ? "Deaktiver" : knastInfo.onpremState === "deactivated" ? "Aktiver" : ""}
        </button>
        {(knastInfo.onpremState === "updating") && <div className="flex flex-row">
          <div className="text-sm" style={{ color: ColorAuxText }}>oppdater</div>
          <Loader size="small" />
        </div>}

        <button className="text-sm text-blue-600 hover:underline" onClick={onConfigureOnprem}
          hidden={!knastInfo.onpremState
            || knastInfo.onpremState === "updating"
          }>
          Konfigurer
        </button>
      </div>
    </div>
  </div>);


  const toggleInternetUrl = (id: string) => {
    const urlItem = knastInfo.internetUrls.items.find((item: any) => item.id === id);
    if (!urlItem) {
      return;
    }
    updateUrlItem.mutateAsync({
      ...urlItem,
      selected: !urlItem.selected
    });
  }

  const UrlList = () => (<div className="min-w-80">
    {
      knastInfo.internetUrls ? knastInfo.internetUrls.items?.length > 0 ?
        knastInfo.internetUrls.items
          .map((urlEntry: any, index: number) => {
            const expires = new Date(urlEntry.expiresAt)
            const expiresIn = expires.getTime() - Date.now()
            const hours = Math.floor(expiresIn / (1000 * 60 * 60));
            const minutes = Math.floor((expiresIn % (1000 * 60 * 60)) / (1000 * 60));
            const durationText = hours > 0 ? `${hours}t ${minutes}m` : `${minutes}m`;

            return (
              knastInfo.internetState === "deactivated" ?
                <div className="pt-2" key={index}>
                  <Checkbox checked={urlEntry.selected} size="small"
                    onChange={() => toggleInternetUrl(urlEntry.id)}
                  > <div className="flex flex-row gap-x-2"><p>{urlEntry.url}</p><p style={{
                    color: ColorAuxText
                  }}>{urlEntry.duration === "01:00:00" ? "1t" : urlEntry.duration === "12:00:00" ? "12t" : "?t"}</p></div></Checkbox>
                </div>
                : <div className="grid grid-cols-[20px_1fr] items-center">
                  {knastInfo.operationalStatus !== "started" ? <IconConnectLightGray />
                    : urlEntry.selected ? new Date(urlEntry.expiresAt) > new Date() ? <IconConnected width={12} /> : <IconConnected width={12} /> : <IconConnectLightGray />}
                  <div key={index} className="flex flex-row gap-x-2 items-center"><p style={{
                    color: urlEntry.selected ? ColorDefaultText : ColorDisabled
                  }}>{urlEntry.url}</p>
                    {urlEntry.selected && new Date(urlEntry.expiresAt) < new Date() ? <p className="text-sm" style={{
                      color: ColorFailed
                    }}>Utløpt</p> : knastInfo.operationalStatus === "started" && urlEntry.selected ? <p className="text-sm" style={{
                      color: ColorSuccessful
                    }}>{durationText}</p> : undefined}</div>
                </div>
            )
          })
        : <div>{"Ikke konfigurert"}</div> : undefined
    }
    <div className="flex flex-row space-x-4 mt-2">
      <div className="flex flex-row items-center space-x-4">
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => knastInfo.internetState === "activated" ? onDeactivateInternet() : knastInfo.internetState === "deactivated" ? onActivateInternet() : undefined}
          disabled={knastInfo.internetState === "updating"}
          hidden={!knastInfo.internetUrls?.items?.length
            || knastInfo.operationalStatus !== "started" || !knastInfo.internetState
            || knastInfo.internetState === "updating"
          }>
          {knastInfo.internetState === "activated" ? "Deaktiver (velg url-er på nytt)" : knastInfo.internetState === "deactivated" ? "Aktiver" : ""}
        </button>
        {(knastInfo.internetState === "updating") && <div className="flex flex-row">
          <div className="text-sm" style={{ color: ColorAuxText }}>oppdater</div>
          <Loader size="small" />
        </div>}

        <button className="text-sm text-blue-600 hover:underline" onClick={() => { onConfigureInternet() }}
          hidden={!knastInfo.internetState || knastInfo.internetState === "updating"}>
          Konfigurer
        </button>
      </div>
    </div>
  </div>)

  return <div className="max-w-180 border-blue-100 border rounded p-4">
    <LocalDevInfo show={showLocalDevInfo} knastInfo={knastInfo} onClose={()=> setShowLocalDevInfo(false)}/>
    <Table>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell colSpan={2} scope="col">
            <h3>Knast - {knastInfo.displayName}</h3>
          </Table.HeaderCell>
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
          <Table.DataCell className="flex flex-rol items-end">{knastInfo.allowSSH ? "Aktivert" : "Deaktivert"}
            {knastInfo.allowSSH && <Link href="#" onClick={()=>setShowLocalDevInfo(true)} className="flex flex-rol ml-2 text-sm">Guide</Link>
            }</Table.DataCell>
        </Table.Row>

        <Table.Row>
          <Table.HeaderCell scope="row">Kostnad</Table.HeaderCell>
          <Table.DataCell>{getKnastDailyCost(knastInfo) || "Ukjent"}</Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Nav datakilder</Table.HeaderCell>
          <Table.DataCell> <OnpremList /> </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Administrerte internettåpninger</Table.HeaderCell>
          <Table.DataCell>{knastInfo.internetUrls.disableGlobalAllowList ? "Deaktivert" : "Aktivert"}
            <button className="pl-4 text-sm text-blue-600 hover:underline" onClick={() => { onConfigureInternet() }}>
              Konfigurer
            </button>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Tilpassede internettåpninger</Table.HeaderCell>
          <Table.DataCell>
            <UrlList />
          </Table.DataCell>
        </Table.Row>
      </Table.Body>
    </Table>
  </div>
}