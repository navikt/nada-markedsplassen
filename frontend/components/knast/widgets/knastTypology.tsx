import { Link, Loader } from "@navikt/ds-react"
import { IconConnected, IconDisconnected, IconInternetOpening, IconNavData, KnastMachine } from "./knastIcons"
import { ColorDisabled, ColorInfo, ColorSuccessful } from "../designTokens"

interface KnastTypologyProps {
    x: number | string
    y: number | string
    className?: string
    onpremHostsNumber: number
    internetOpeningsNumber: number
    onpremState: string
    internetConnected: boolean
    showConnectivity: boolean
    onConfigureOnprem: () => void
    onConfigureInternet: () => void
    onConnectOnprem: () => void
    onConnectInternet: () => void
    onDisconnectOnprem: () => void
    onDisconnectInternet: () => void
}

export const KnastTypology = ({ x, y, className, onpremHostsNumber, internetOpeningsNumber, onpremState, internetConnected, showConnectivity, onConfigureOnprem, onConfigureInternet, onConnectOnprem, onConnectInternet, onDisconnectOnprem, onDisconnectInternet }: KnastTypologyProps) => {
    const navdataColor = onpremState === "activated" ? ColorInfo : ColorDisabled
    const internetColor = internetConnected ? ColorInfo : ColorDisabled

    console.log(onpremState)
    return <div className={`absolute ${className}`} style={{ left: x, top: y }}>
        <div className={"relative"}>
            <div className={"absolute"} style={{ left: 10, top: 10 }}>
                <KnastMachine />
            </div>
            {showConnectivity && <div>
                <div className="absolute w-40 h-px" style={{ left: 240, top: 55, backgroundColor: navdataColor }}></div>
                <button className="absolute" style={{ left: 310, top: 47 }}
                    onClick={onpremState === "activated" ? onDisconnectOnprem : onConnectOnprem}
                    disabled={onpremState === "deactivating" || onpremState === "activating"}
                >
                    {onpremState === "deactivated" ? <IconDisconnected width={16} height={16} /> : onpremState === "activated" ? <IconConnected width={16} height={16} />
                        : <Loader size="small" className="mb-2" />}
                </button>

                <div className={"absolute flex flex-row gap-2 items-center"} style={{ left: 410, top: 30 }}>
                    <IconNavData state={onpremState === "activated" ? "normal" : "grayed"} />
                    <div className="grid grid-rows-3">
                        <div className="text-m w-100 text-[#0067C5]">{`Nav data (${onpremHostsNumber} kilder)`} </div>
                        <Link className="text-sm" href="#" onClick={onConfigureOnprem}>Konfigurer</Link>
                        <Link className="text-sm" hidden={onpremState !== "activated" && onpremState !== "deactivated"} href="#" onClick={onpremState === "activated" ? onDisconnectOnprem : onConnectOnprem}>{onpremState === "activated" ? "Deaktiver" : "Aktiver"}</Link>
                    </div>
                </div>


                <div className="absolute w-40 h-px" style={{ left: 240, top: 155, backgroundColor: internetColor }}></div>
                <button className="absolute" style={{ left: 310, top: 147 }} onClick={internetConnected ? onDisconnectInternet : onConnectInternet}>
                    {!internetConnected ? <IconDisconnected width={16} height={16} /> : <IconConnected width={16} height={16} />}
                </button>
                <div className={"absolute flex flex-row gap-2"} style={{ left: 410, top: 130 }}>
                    <IconInternetOpening />
                    <div className="flex flex-col">
                        <div className="text-m w-100 text-[#0067C5]">Internett {!internetOpeningsNumber && "(0 host)"} </div>
                        <Link className="text-sm" href="#" onClick={onConfigureInternet}>Configure</Link>
                    </div>
                </div>
            </div>
            }
        </div>
    </div>
}