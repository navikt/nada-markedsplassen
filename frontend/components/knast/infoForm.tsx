import { ChevronDownIcon, ChevronUpIcon, CircleSlashIcon, ExclamationmarkTriangleIcon, ExternalLinkFillIcon, ExternalLinkIcon, InformationIcon, InformationSquareFillIcon } from "@navikt/aksel-icons";
import { Alert, Checkbox, Loader, Table, Tooltip } from "@navikt/ds-react";
import Link from "next/link";
import React, { use } from "react";
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
import { UrlItem } from "./widgets/urlItem";
import { InfoLink } from "./widgets/infoLink";
import { LogViewer } from "./widgets/logViewer";
import { useOnpremMapping } from "../onpremmapping/queries";

type InfoFormProps = {
  knastInfo: any
  operationalStatus?: string
  onActivateOnprem: () => void
  onActivateInternet: () => void
  onDeactivateOnPrem: () => void
  onDeactivateInternet: () => void
  onConfigureOnprem: () => void
  onConfigureInternet: () => void
  onConfigureSSH: () => void
  onShowLogs?: () => void
}

const operationStatusText = new Map<string, string>([
  ["started", "Kjører"],
  ["stopped", "Stoppet"],
  ["starting", "Starter"],
  ["stopping", "Stopper"],
])

export const InfoForm = ({ knastInfo, operationalStatus, onActivateOnprem,  onActivateInternet
  , onDeactivateOnPrem, onDeactivateInternet, onConfigureOnprem, onConfigureInternet, onConfigureSSH, onShowLogs }: InfoFormProps) => {
  const [showAllDataSources, setShowAllDataSources] = React.useState(false);
  const updateUrlItem = useUpdateWorkstationURLListItemForIdent();
  const [showLocalDevInfo, setShowLocalDevInfo] = React.useState(false);
  const [selectedItems, setSelectedItems] = React.useState<string[] | undefined>(undefined);
  const onpremMapping = useOnpremMapping();

  const allOnpremHosts = [
    ...knastInfo?.workstationOnpremMapping,
    ...knastInfo?.effectiveTags?.tags?.filter((it: any) => !knastInfo?.workstationOnpremMapping?.some((mapping: any) => mapping.host === it.namespacedTagKey?.split("/").pop())).map((it: any) => {
      return { host: it.namespacedTagKey?.split("/").pop(), isDVHSource: false }
    }) || []
  ]
  const showActivateOnprem = knastInfo.onpremState !== "updating"
    && (knastInfo.effectiveTags?.tags?.length || 0) < allOnpremHosts.filter(it => !it.isDVHSource || !knastInfo.allowSSH).length && knastInfo.operationalStatus === "started";
  const showDeactivateOnprem = knastInfo.onpremState !== "updating" && knastInfo.effectiveTags?.tags?.length && knastInfo.operationalStatus === "started";
  const showActivateInternet = knastInfo.internetUrls?.items?.length && knastInfo.operationalStatus === "started"
    && knastInfo.internetUrls?.items?.some((it: any) => it.selected && new Date(it.expiresAt) < new Date());
  const showDeactivateInternet = knastInfo.internetUrls?.items?.length && knastInfo.operationalStatus === "started"
    && knastInfo.internetUrls?.items?.some((it: any) => it.selected && new Date(it.expiresAt) > new Date());
  const showRefreshInternet = knastInfo.internetState === "activated"
    && knastInfo.internetUrls?.items?.length && knastInfo.operationalStatus === "started";
  const backendSelectedItems = () => knastInfo.internetUrls?.items?.filter((it: any) => it.selected).map((it: any) => it.id) || [];

  React.useEffect(() => {
    setSelectedItems(backendSelectedItems());
  }, [knastInfo.internetUrls]);

  const getOnpremHostDisplayName = (host: any) => {
    const fullHost = Object.values(onpremMapping.data?.hosts ?? {}).flat().find((it: any) => it.Host === host);
    return fullHost ? fullHost.Name ? (fullHost.Name !== host ? `${fullHost.Name} (${host})` : fullHost.Name) : host : host;
  }
  const OnpremList = () => (<div>
    {
      allOnpremHosts.length > 0 ?
        allOnpremHosts.slice(0, showAllDataSources ? allOnpremHosts.length : 5)
          .map((mapping: any, index: number) => (
            <div key={index} className="grid grid-cols-[20px_1fr] items-center">
              {knastInfo.operationalStatus !== "started" ? <IconConnectLightGray />
                : knastInfo.effectiveTags?.tags?.find((tag: any) => tag.namespacedTagKey?.split("/").pop() === mapping.host)
                  ? <IconConnected width={16} />
                  : mapping.isDVHSource && knastInfo.allowSSH ? <IconConnectLightGray /> : <IconDisconnected width={16} />}
              <Tooltip hidden={(!mapping.isDVHSource || !knastInfo.allowSSH) && (operationalStatus === "started")}
                content={mapping.isDVHSource && knastInfo.allowSSH ? "Denne kilden er en DVH-kilde og kan ikke nås når SSH er aktivert" : operationalStatus !== "started" ? "Du kan ikke aktivere tilkoblinger når knast ikke er startet" : ""}>
                <div className="ml-2" key={index} style={{
                  color: mapping.isDVHSource && knastInfo.allowSSH ? ColorDisabled : ColorDefaultText
                }}>{getOnpremHostDisplayName(mapping.host)}</div>
              </Tooltip>
            </div>
          ))
        : <div>{"Ikke konfigurert"}</div>
    }
    <div className="flex flex-row space-x-4 mt-2">
      {knastInfo.workstationOnpremMapping && knastInfo.workstationOnpremMapping.length > 10 &&
        <button className="text-sm text-blue-600 hover:underline" onClick={() => setShowAllDataSources(!showAllDataSources)}>
          {showAllDataSources ? "Vis færre" : `Vis alle (${knastInfo.workstationOnpremMapping.length})`}
        </button>
      }
      <div className="flex flex-row items-center space-x-4">
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => onActivateOnprem()}
          hidden={!showActivateOnprem}>
          Aktiver
        </button>
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => onDeactivateOnPrem()}
          hidden={!showDeactivateOnprem}>
          Deaktiver
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
    setSelectedItems(selectedItems ? (selectedItems.includes(id) ? selectedItems.filter(it => it !== id) : [...selectedItems, id]) : [id]);
    updateUrlItem.mutateAsync({
      ...urlItem,
      selected: !urlItem.selected
    });
  }

  const UrlList = () => (<div className="w-80">
    {
      knastInfo.internetUrls ? knastInfo.internetUrls.items?.length > 0 ?
        knastInfo.internetUrls.items
          .map((urlEntry: any, index: number) => {
            return (<div className="wrap-break-word" key={index}>{
              knastInfo.operationalStatus !== "started" ?
                <UrlItem item={urlEntry} style="status" status="unavailable" />
                : urlEntry.selected && new Date(urlEntry.expiresAt) > new Date() ?
                  <UrlItem item={urlEntry} style="status" status="connected" />

                  : knastInfo.internetState !== "updating" ?
                    <UrlItem item={urlEntry} style="pick" status={"pickable"} selectedItems={selectedItems} onToggle={() => toggleInternetUrl(urlEntry.id)} />
                    : <UrlItem item={urlEntry} style="pick" status={"disabled"} selectedItems={selectedItems} onToggle={() => toggleInternetUrl(urlEntry.id)} />
            }
            </div>)
          })
        : <div>{"Ikke konfigurert"}</div> : undefined
    }
    <div className="flex flex-row space-x-4 mt-2">
      <div className="flex flex-row items-center space-x-4">
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => onActivateInternet()}
          hidden={!showActivateInternet}>
          Aktiver
        </button>
        <button className="text-sm text-blue-600 hover:underline justify-end"
          onClick={() => onDeactivateInternet()}
          hidden={!showDeactivateInternet}>
          Deaktiver
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

  return <div className="w-180 border-blue-100 border p-4">
    {!!knastInfo.blockedUrls.length
      && <Alert variant="warning" className="mb-4">
        {knastInfo.blockedUrls.length} {knastInfo.blockedUrls.length > 1 ? "URL-er" : "URL"} ble blokkert i løpet av den siste timen, se <Link href="#" onClick={onShowLogs}>logger</Link>
      </Alert>}
    <LocalDevInfo show={showLocalDevInfo} knastInfo={knastInfo} onClose={() => setShowLocalDevInfo(false)} />
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
            <div className="flex flex-row">
              {operationStatusText.get(operationalStatus ?? "") || "Ukjent"}
              {operationalStatus === "started" &&
                <InfoLink className="ml-2 text-sm" caption={"Avstengningspolicy"} content={<div>
                  En kjørende Knast vil <strong>stenges etter 2 timer uten aktivitet</strong>. Den vil også ha en hard
                  grense på <strong>12 timer</strong> for hver økt. Dette er for å sikre at ressursene i skyen ikke
                  blir brukt unødvendig, og ha muligheten til å kjøre sikkerthetsoppdateringer.</div>}
                />}
            </div>
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
            }}>{knastInfo.machineTypeInfo && `${knastInfo.machineTypeInfo.vCPU} vCPU, ${knastInfo.machineTypeInfo.memoryGB} GB RAM, ${getKnastDailyCost(knastInfo)}`}</div>
          </Table.DataCell>
        </Table.Row>

        <Table.Row>
          <Table.HeaderCell scope="row">Lokal dev (SSH)</Table.HeaderCell>
          <Table.DataCell className="flex flex-rol items-end">{knastInfo.allowSSH ? "Aktivert" : "Deaktivert"}
            <Link href="#" onClick={() => onConfigureSSH()} className="flex flex-rol ml-2 text-sm">Konfigurer</Link>
            {knastInfo.allowSSH && <Link href="#" onClick={() => setShowLocalDevInfo(true)} className="flex flex-rol ml-2 text-sm">Guide</Link>
            }</Table.DataCell>

        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Nav datakilder</Table.HeaderCell>
          <Table.DataCell> <OnpremList /> </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Administrerte internettåpninger</Table.HeaderCell>
          <Table.DataCell>{knastInfo.internetUrls?.disableGlobalAllowList ? "Deaktivert" : "Aktivert"}
            <button className="pl-4 text-sm text-blue-600 hover:underline" onClick={() => { onConfigureInternet() }}>
              Konfigurer
            </button>
          </Table.DataCell>
        </Table.Row>
        <Table.Row>
          <Table.HeaderCell scope="row">Tidsbegrensede internettåpninger</Table.HeaderCell>
          <Table.DataCell>
            <UrlList />
          </Table.DataCell>
        </Table.Row>
      </Table.Body>
    </Table>
  </div>
}