import { Link, Loader, Tooltip } from "@navikt/ds-react"
import { IconConnected, IconDisconnected, IconInternetOpening, IconNavData, KnastMachine } from "./knastIcons"
import { ColorAuxText, ColorDisabled, ColorInfo, ColorSuccessful } from "../designTokens"

interface KnastTypologyProps {
    x: number | string
    y: number | string
    className?: string
    onpremHostsNumber: number
    internetOpeningsNumber: number
    onpremState: string
    internetState: string
    showConnectivity: boolean
    onpremConnectedNumber?: number
    onConfigureOnprem: () => void
    onConfigureInternet: () => void
    onConnectOnprem: () => void
    onConnectInternet: () => void
    onDisconnectOnprem: () => void
    onDisconnectInternet: () => void
}

export const KnastTypology = ({ x, y, className, onpremHostsNumber, internetOpeningsNumber
    , onpremConnectedNumber, onpremState, showConnectivity, internetState
    , onConfigureOnprem, onConfigureInternet, onConnectOnprem, onConnectInternet, onDisconnectOnprem, onDisconnectInternet }: KnastTypologyProps) => {
    const navdataColor = onpremState === "activated" ? ColorInfo : ColorDisabled
    const internetColor = internetState === "activated" ? ColorInfo : ColorDisabled
    const disableConnectOnprem = !onpremHostsNumber && onpremState != "activated" || onpremState !== "activated" && onpremState !== "deactivated"
    const disableConnectInternet = !internetOpeningsNumber && internetState != "activated" || internetState !== "activated" && internetState !== "deactivated"

    return <div className={`absolute ${className}`} style={{ left: x, top: y }}>
        <div className={"relative"}>
            <div className={"absolute"} style={{ left: 10, top: 10 }}>
                <KnastMachine />
            </div>
            {showConnectivity && <div>
                <div className="absolute w-40 h-px" style={{ left: 240, top: 55, backgroundColor: navdataColor }}></div>
                <Tooltip hidden={onpremState === "updating"} content={onpremState === "deactivated" ? "koble til Nav data" : onpremState === "activated" ? "koble fra Nav data" : "Oppdaterer tilkobling til Nav data"}>
                    <button className="absolute" style={{ left: 310, top: 47 }}
                        onClick={onpremState === "activated" ? onDisconnectOnprem : onConnectOnprem}
                        disabled={disableConnectOnprem}
                    >
                        {onpremState === "deactivated" ? <IconDisconnected width={16} height={16} /> : onpremState === "activated" ? <IconConnected width={16} height={16} />
                            : <Loader size="small" className="mb-2" />}
                    </button>

                </Tooltip>

                <div className={"absolute flex flex-row gap-2 items-center"} style={{ left: 410, top: 30 }}>
                    <IconNavData state={onpremState === "activated" ? "normal" : "grayed"} />
                    <div className="grid grid-rows-3">
                        <div className="text-m w-100 text-[#0067C5]">{`Nav data${onpremConnectedNumber ? ` (${onpremConnectedNumber} kilder)` : ""}`} </div>
                        <Link className="text-sm" hidden={onpremState === "updating"} href="#" onClick={onConfigureOnprem}>Konfigurer</Link>
                        <Link className="text-sm" hidden={disableConnectOnprem} href="#" onClick={onpremState === "activated" ? onDisconnectOnprem : onConnectOnprem}>{onpremState === "activated" ? "Deaktiver" : "Aktiver"}</Link>
                        {onpremState === "updating" && <div className="text-sm italic" style={{ color: ColorAuxText }}>Oppdater</div>}
                    </div>
                </div>


                <div className="absolute w-40 h-px" style={{ left: 240, top: 155, backgroundColor: internetColor }}></div>
                <Tooltip hidden={internetState === "updating"} content={internetState === "deactivated" ? "koble til Internett" : internetState === "activated" ? "koble fra Internett" : "Oppdaterer tilkobling til Internett"}>
                    <button className="absolute" style={{ left: 310, top: 147 }}
                        onClick={internetState === "activated" ? onDisconnectInternet : onConnectInternet}
                        disabled={disableConnectInternet}>
                        {internetState === "deactivated" ? <IconDisconnected width={16} height={16} /> : internetState === "activated" ? <IconConnected width={16} height={16} />
                            : <Loader size="small" className="mb-2" />}
                    </button>
                </Tooltip>
                <div className={"absolute flex flex-row gap-2"} style={{ left: 410, top: 130 }}>
                    <IconInternetOpening state={internetState === "activated" ? "normal" : "grayed"} />
                    <div className="flex flex-col">
                        <div className="text-m w-100 text-[#0067C5]">Internett (tilpassede url-er)</div>
                        <Link className="text-sm" hidden={internetState === "updating"} href="#" onClick={onConfigureInternet}>Konfigurer</Link>
                        <Link className="text-sm" hidden={disableConnectInternet} href="#" onClick={internetState === "activated" ? onDisconnectInternet : onConnectInternet}>{internetState === "activated" ? "Deaktiver" : "Aktiver"}</Link>
                        {internetState === "updating" && <div className="text-sm italic" style={{
                            color: ColorAuxText
                        }}>Oppdater</div>}
                    </div>
                </div>
            </div>
            }
        </div>
    </div>
}