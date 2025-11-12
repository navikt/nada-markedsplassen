import { Link } from "@navikt/ds-react";
import { WorkstationOutput } from "../../lib/rest/generatedDto";
import { ColorAuxText, ColorInfoText } from "./designTokens";
import { useAnimatePlayButton } from "./widgets/animatePlayButton";
import { IconButton } from "./widgets/iconButton";
import { IconGear, IconOpenedBook } from "./widgets/knastIcons";
import { KnastTypology } from "./widgets/knastTypology";
import { KnastDisplay } from "./widgets/knastDisplay";

interface ControlPanelProps {
    knastInfo: any
    onStartKnast: () => void
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
    const { PlayButton, operationalStatus } = useAnimatePlayButton(knast)
    return {
        operationalStatus,
        ControlPanel: ({ knastInfo, onStartKnast, onStopKnast, onSettings, onActivateOnprem, onActivateInternet, onDeactivateOnPrem, onDeactivateInternet, onConfigureOnprem, onConfigureInternet }: ControlPanelProps) => {
            return <div className="relative h-80">
                <KnastTypology x={0} y={22}
                    showConnectivity={operationalStatus === "started"}
                    onpremConnectedNumber={knastInfo.effectiveTags?.tags?.length || 0}
                    internetOpeningsNumber={knastInfo.internetUrls?.items?.length || 0}
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
                <PlayButton onButtonStart={onStartKnast} onButtonStop={onStopKnast} x={40} y={190} />
                <IconButton x={200} y={190} ariaLabel="Settings" tooltip="Settings" onClick={onSettings}>
                    <IconGear />
                </IconButton>
                <IconButton x={160} y={190} ariaLabel="Documentation" tooltip="Documentation" onClick={() => alert("Documentation")}>
                    <IconOpenedBook />
                </IconButton>
            </div>
        }
    }
}