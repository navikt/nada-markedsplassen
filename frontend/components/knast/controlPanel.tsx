import React, { useState } from "react";
import { ConnectButton, DatasourcesConnectedIcon, DatasourcesDisconnectedIcon, DocumentationIcon, InternetAccessIcon, KnastMachine, OnPremSourcesIcon, SettingsIcon, StartButton, StopButton } from "./knastIcons";
import { Button, Link, Loader, Tag, Tooltip } from "@navikt/ds-react";

type Hotspot = {
    x: number;
    y: number;
    label?: string;
    onClick?: () => void;
    render?: React.ReactNode;
    ariaLabel?: string;
    tooltip: string;
};

type KnastInteractiveIconProps = {
    env: string;
    operationalStatus: string;
    message: string;
    connectivity: string;
    hotspots: Hotspot[];
    className?: string;
};


const KnastInteractiveIcon = ({
    hotspots,
    className,
    env,
    operationalStatus: operationalStatus,
    connectivity,
    message,
}: KnastInteractiveIconProps) => {
    return (
        <div
            className={className}
            style={{
                position: "relative",
                width: "240px",
                display: "inline-block",
                lineHeight: 0,
            }}
        >
            <KnastMachine />
            <div className="text-blue-600" style={{
                position: "absolute",
                left: "10px",
                top: "30px",
            }}>
                {env}
            </div>
            <div className={operationalStatus == "Stopped"? "text-gray-600": operationalStatus == "Starting..."|| operationalStatus =="Stopping..."? "text-black": "text-green-400"} style={{
                position: "absolute",
                left: "140px",
                top: "30px",
            }}>
                {operationalStatus}
            </div>

            {operationalStatus == "Running" && <div style={{
                position: "absolute",
                left: "80px",
                top: "80px",
            }}>
                <Link href="#">Open in browser</Link>
                <Link href="#" className="mt-8 flex flex-rol">Local dev?</Link>
                </div>}
            {operationalStatus=="Running" && <div className={connectivity? "text-green-400":"text-gray-600"} style={{
                position: "absolute",
                left: "10px",
                top: "100px",
            }}>
                {connectivity=="Connected"? <DatasourcesConnectedIcon/>: connectivity == "Disconnected"? <DatasourcesDisconnectedIcon/>: <Loader />}
            </div>}
            {hotspots.map((h, i) => {
                const aria = h.ariaLabel || `Hotspot ${i + 1}`;
                return (
                    <Tooltip content={h.tooltip} key={i}>
                        <Button
                            type="button"
                            aria-label={aria}
                            onClick={h.onClick}
                            tabIndex={i}
                            style={{
                                position: "absolute",
                                left: `${h.x}%`,
                                top: `${h.y}%`,
                                transform: "translate(-50%, -50%)",
                                padding: "0.5rem 0.75rem",
                                borderRadius: "9999px",
                                border: "none",
                                cursor: "pointer",
                                boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
                                background: "white",
                                font: "inherit",
                            }}
                        >
                            {h.render ?? h.label ?? "•"}
                        </Button>

                    </Tooltip>
                );
            })}
        </div>
    );
}

export const useKnastControlPanel = () => {
    const [operationalStatus, SetOperationalStatus] = useState("Stopped")
    const [environment, setEnvironment] = useState("VS Code")
    const [connectivity, setConnectivity] = useState("Disconnected")
    const [logs, setLogs] = useState<string[]>([] as string[])
    const [activeForm, setActiveForm] = useState<"status" | "settings" | "sources" | "internet">("status")
    const getFormattedTimeStamp = () => {
        const now = new Date();
        return now.toLocaleDateString("no-NO", {
            month: "short",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
        });
    }

    return {
        operationalStatus,
        environment,
        connectivity,
        activeForm,
        setActiveForm,
        logs,
        component: ()=> <div>
        <section className={`flex flex-rol`}>
            <KnastInteractiveIcon
                className=""
                env={environment}
                operationalStatus={operationalStatus}
                message={""}
                connectivity={connectivity}
                hotspots={[
                    {
                        x: 15,
                        y: 72,
                        render: <StartButton />,
                        onClick: () => {
                            SetOperationalStatus("Starting...")
                            setTimeout(()=>SetOperationalStatus("Running"), 3000)
                            setTimeout(()=>setLogs([`${getFormattedTimeStamp()} - Knast started`, ...logs]), 3000)
                        },
                        ariaLabel: "Start Knast",
                        tooltip: "Start your Knast"
                    },
                    {
                        x: 35,
                        y: 72,
                        render: <StopButton />,
                        onClick: () => {
                            SetOperationalStatus("Stopping...")
                            setTimeout(()=>SetOperationalStatus("Stopped"), 3000)
                            setTimeout(()=>setLogs([`${getFormattedTimeStamp()} - Knast stopped`, ...logs]), 3000)
                        },
                        ariaLabel: "Stop Knast",
                        tooltip: "Stop your Knast"
                    },
                    {
                        x: 80,
                        y: 72,
                        render: <ConnectButton />,
                        onClick: () => {
                            setConnectivity("Updating...")
                            setTimeout(()=>setConnectivity(connectivity=="Connected"?"Disconnected":"Connected"), 3000)
                            setTimeout(()=>setLogs([`${getFormattedTimeStamp()} - Data sources connected`, ...logs]), 3000)
                        },
                        ariaLabel: "Connect to data sources",
                        tooltip: "Connect to selected data sources"
                    },
                ]}
            />


            <div className="flex-1 flex flex-col pt-5 pb-5">
                <div className="flex flex-col gap-2">

                    <div className="flex flex-row">
                        <SettingsIcon />
                        <Tooltip content="Change knast settings">
                            <Link href="#" onClick={() => setActiveForm("settings")}>Settings</Link>
                        </Tooltip>
                    </div>
                    <div className="flex flex-row"><OnPremSourcesIcon />
                        <Tooltip content="Select on-prem data sources for knast">
                            <Link href="#" onClick={() => setActiveForm("sources")}>On-Prem sources</Link>

                        </Tooltip>
                    </div>
                    <div className="flex flex-row"><InternetAccessIcon />
                        <Tooltip content="Configure allowed internet hosts for knast">
                            <Link href="#" onClick={() => setActiveForm("internet")}>Internet gateway</Link>

                        </Tooltip>
                    </div>

                </div>

                <div className="mt-auto">
                    <div className="flex flex-row"><DocumentationIcon />
                        <Tooltip content="Everything about knast">
                            <Link href="#">Documentation</Link>
                        </Tooltip>
                    </div>
                </div>
            </div>
        </section>
    </div>
    }
}
