import { Link } from "@navikt/ds-react";
import { WorkstationOutput } from "../../lib/rest/generatedDto";
import { ColorAuxText, ColorInfoText } from "./designTokens";
import { useAnimatePlayButton } from "./widgets/animatePlayButton";
import { IconButton } from "./widgets/iconButton";
import { IconGear, IconOpenedBook } from "./widgets/knastIcons";
import { KnastTypology } from "./widgets/knastTypology";
import { KnastDisplay } from "./widgets/knastDisplay";

interface ControlPanelProps {
    knastInfo: WorkstationOutput
    onStartKnast: () => void
    onStopKnast: () => void
    onSettings: () => void
    onConnectOnprem: () => void
    onConnectInternet: () => void
    onDisconnectOnprem: () => void
    onDisconnectInternet: () => void
    onConfigureOnprem: () => void
    onConfigureInternet: () => void
}

export const useControlPanel = (knastInfo: WorkstationOutput | undefined) => {
    const { PlayButton, operationalStatus } = useAnimatePlayButton(knastInfo)
    return {
        operationalStatus,
        ControlPanel: ({ knastInfo, onStartKnast, onStopKnast, onSettings, onConnectOnprem, onConnectInternet, onDisconnectOnprem, onDisconnectInternet, onConfigureOnprem, onConfigureInternet }: ControlPanelProps) => {
            return <div className="relative h-80">
                <KnastTypology x={0} y={22}
                    showConnectivity={operationalStatus === "Started"}
                    onpremHostsNumber={0} internetOpeningsNumber={3} onpremConnected={true} internetConnected={false}
                    onConfigureInternet={onConfigureInternet}
                    onConfigureOnprem={onConfigureOnprem}
                    onConnectInternet={onConnectInternet}
                    onConnectOnprem={onConnectOnprem}
                    onDisconnectInternet={onDisconnectInternet}
                    onDisconnectOnprem={onDisconnectOnprem}
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