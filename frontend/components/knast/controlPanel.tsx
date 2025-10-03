import React from "react";
import { ConnectButton, DocumentationIcon, InternetAccessIcon, KnastMachine, OnPremSourcesIcon, SettingsIcon, StartButton, StopButton } from "./knastIcons";
import { Button, Link, Tooltip } from "@navikt/ds-react";

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
    src: string;
    alt: string;
    hotspots: Hotspot[];
    className?: string;
};

const KnastInteractiveIcon = ({
    hotspots,
    className,
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

const KnastControlPanel = () => {
    return <div>
        <section className={`flex flex-rol`}>
            <KnastInteractiveIcon
                src="/knast/control-panel.png"
                alt="Knast Control Panel"
                className=""
                hotspots={[
                    {
                        x: 15,
                        y: 72,
                        render: <StartButton />,
                        onClick: () => alert("Start knast requested"),
                        ariaLabel: "Start Knast",
                        tooltip: "Start your Knast"
                    },
                    {
                        x: 35,
                        y: 72,
                        render: <StopButton />,
                        onClick: () => alert("Stop knast requested"),
                        ariaLabel: "Stop Knast",
                        tooltip: "Stop your Knast"
                    },
                    {
                        x: 80,
                        y: 72,
                        render: <ConnectButton />,
                        onClick: () => alert("Connect to on-prem sources requested"),
                        ariaLabel: "Connect to on-prem sources",
                        tooltip: "Connect to selected on-prem sources"
                    },
                ]}
            />


            <div className="flex-1 flex flex-col pt-5 pb-5">
                <div className="flex flex-col gap-2">

                    <div className="flex flex-row">
                        <SettingsIcon />
                        <Tooltip content="Change knast settings">
                            <Link href="#">Settings</Link>
                        </Tooltip>
                    </div>
                    <div className="flex flex-row"><OnPremSourcesIcon />
                        <Tooltip content="Select on-prem data sources for knast">
                            <Link href="#">On-Prem sources</Link>

                        </Tooltip>
                    </div>
                    <div className="flex flex-row"><InternetAccessIcon />
                        <Tooltip content="Configure allowed internet hosts for knast">
                            <Link href="#">Internet gateway</Link>

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

export default KnastControlPanel