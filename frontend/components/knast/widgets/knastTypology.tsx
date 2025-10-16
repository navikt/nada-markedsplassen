import { Link } from "@navikt/ds-react"
import { IconConnected, IconDisconnected, IconInternetOpening, IconNavData, KnastMachine } from "./knastIcons"

interface KnastTypologyProps {
    x: number | string
    y: number | string
    className?: string
    onpremHostsNumber: number
    internetOpeningsNumber: number
    onpremConnected: boolean
    internetConnected: boolean
    showConnectivity: boolean
    onConfigureOnprem: () => void
    onConfigureInternet: () => void
    onConnectOnprem: () => void
    onConnectInternet: () => void
    onDisconnectOnprem: () => void
    onDisconnectInternet: () => void
}

export const KnastTypology = ({ x, y, className, onpremHostsNumber, internetOpeningsNumber, onpremConnected, internetConnected, showConnectivity, onConfigureOnprem, onConfigureInternet, onConnectOnprem, onConnectInternet, onDisconnectOnprem, onDisconnectInternet }: KnastTypologyProps) => {
    const navdataLineColor = onpremConnected ? "#0067C5" : "#A6A6A6"
    const internetLineColor = internetConnected ? "#0067C5" : "#A6A6A6"

    return <div className={`absolute ${className}`} style={{ left: x, top: y }}>
        <div className={"relative"}>
            <div className={"absolute"} style={{ left: 10, top: 10 }}>
                <KnastMachine />
            </div>
            {showConnectivity && <div>
                <div className="absolute w-40 h-px" style={{ left: 240, top: 55, backgroundColor: navdataLineColor }}></div>
                <button className="absolute" style={{ left: 310, top: 47 }} onClick={onpremConnected ? onDisconnectOnprem : onConnectOnprem}>
                    {!onpremConnected ? <IconDisconnected width={16} height={16} /> : <IconConnected width={16} height={16} />}
                </button>
                <div className={"absolute flex flex-row gap-2 items-center"} style={{ left: 410, top: 30 }}>
                    <IconNavData />
                    <div className="flex flex-col">
                        <div className="text-m w-100 text-[#0067C5]">Nav data {!onpremHostsNumber && "(0 host)"} </div>
                        <Link className="text-sm" href="#" onClick={onConfigureOnprem}>Configure</Link>
                        {!!onpremHostsNumber && <Link className="text-sm" href="#" onClick={onpremConnected ? onDisconnectOnprem : onConnectOnprem}>{onpremConnected ? "Disconnect" : "Connect"}</Link>}
                    </div>
                </div>
                <div className="absolute w-40 h-px" style={{ left: 240, top: 155, backgroundColor: internetLineColor }}></div>
                <button className="absolute" style={{ left: 310, top: 147 }} onClick={internetConnected ? onDisconnectInternet : onConnectInternet}>
                    {!internetConnected ? <IconDisconnected width={16} height={16} /> : <IconConnected width={16} height={16} />}
                </button>
                <div className={"absolute flex flex-row gap-2"} style={{ left: 410, top: 130 }}>
                    <IconInternetOpening />
                    <div className="flex flex-col">
                        <div className="text-m w-100 text-[#0067C5]">Internet {!internetOpeningsNumber && "(0 host)"} </div>
                        <Link className="text-sm" href="#" onClick={onConfigureInternet}>Configure</Link>
                        {!!internetOpeningsNumber && <Link className="text-sm" href="#" onClick={internetConnected ? onDisconnectInternet : onConnectInternet}>{internetConnected ? "Disconnect" : "Connect"}</Link>}
                    </div>
                </div>
            </div>
            }
        </div>
    </div>
}