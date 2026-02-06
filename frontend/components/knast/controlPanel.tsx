import { Link, Popover, Tooltip } from "@navikt/ds-react";
import { WorkstationOutput } from "../../lib/rest/generatedDto";
import { ColorAuxText, ColorInfoText } from "./designTokens";
import { useAnimatePlayButton } from "./widgets/animatePlayButton";
import { IconButton } from "./widgets/iconButton";
import { IconGear, IconOpenedBook } from "./widgets/knastIcons";
import { KnastTypology } from "./widgets/knastTypology";
import { KnastDisplay } from "./widgets/knastDisplay";
import React, { useEffect } from "react";

interface ControlPanelProps {
    knastInfo: any
    onStartKnast: () => void
    onRestartKnast: () => void
    onStopKnast: () => void
    onSettings: () => void
    onActivateOnprem: () => void
    onActivateInternet: () => void
    onDeactivateOnPrem: () => void
    onDeactivateInternet: () => void
    onConfigureOnprem: () => void
    onConfigureInternet: () => void
}

export const useControlPanel = (knast: WorkstationOutput | undefined) => {
    const { PlayButton, RestartButton,operationalStatus } = useAnimatePlayButton(knast)
    return {
        operationalStatus,
        ControlPanel: ({ knastInfo, onStartKnast, onRestartKnast, onStopKnast, onSettings, onActivateOnprem, onActivateInternet, onDeactivateOnPrem, onDeactivateInternet, onConfigureOnprem, onConfigureInternet }: ControlPanelProps) => {
            return <div className="relative h-70">
                <KnastTypology x={0} y={22}
                    showConnectivity={operationalStatus === "started"}
                    onpremConnectedNumber={knastInfo.effectiveTags?.tags?.length || 0}
                    internetOpeningsNumber={knastInfo.internetUrls?.items?.filter((it: any) => it.selected).length || 0}
                    onpremHostsNumber={knastInfo.validOnpremHosts?.length || 0}
                    onpremState={knastInfo.onpremState}
                    internetState={knastInfo.internetState}
                    onConfigureInternet={onConfigureInternet}
                    onConfigureOnprem={onConfigureOnprem}
                    onConnectInternet={onActivateInternet}
                    onConnectOnprem={onActivateOnprem}
                    onDisconnectInternet={onDeactivateInternet}
                    onDisconnectOnprem={onDeactivateOnPrem}
                />
                <KnastDisplay operationalStatus={operationalStatus} knastInfo={knastInfo} />
                <PlayButton onButtonStart={onStartKnast} onButtonStop={onStopKnast} x={50} y={197} />
                {operationalStatus === "started" && <RestartButton onButtonRestart={onRestartKnast} x={78} y={180} />}
                <IconButton x={218} y={197} ariaLabel="Settings" tooltip="Innstillinger" onClick={onSettings}>
                    <IconGear />
                </IconButton>
            </div>
        }
    }
}