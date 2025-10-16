import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../lib/rest/generatedDto";
import { ColorDefaultTextInvert, ColorDisabled, ColorInProgress, ColorSuccessful } from "./designTokens";
import { useAnimatePlayButton } from "./widgets/animatePlayButton";
import { IconButton } from "./widgets/iconButton";
import { IconGear, IconOpenedBook } from "./widgets/knastIcons";
import { KnastTypology } from "./widgets/knastTypology";


interface OperationalStatusProps {
    operationalStatus: string
}

const OperationalStatus = ({operationalStatus}: OperationalStatusProps) => {
    const bannerX = 22
    const bannerY = 40
    const bannerWidth = 200
    const bannerHeight = 30
    return <div>
    {operationalStatus === "Started" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorSuccessful,
            color: ColorDefaultTextInvert,
        }}>Kjører</div>}
        {operationalStatus === "Stopped" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorDisabled,
            color: ColorDefaultTextInvert,
        }}>Stoppet</div>}
        {operationalStatus === "Starting" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorInProgress,
            color: ColorDefaultTextInvert,
        }}>Starter</div>}
        {operationalStatus === "Stopping" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorInProgress,
            color: ColorDefaultTextInvert,
        }}>Stopper</div>}
    </div>

 
}

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

export const ControlPanel = ({ knastInfo, onStartKnast, onStopKnast, onSettings, onConnectOnprem, onConnectInternet, onDisconnectOnprem, onDisconnectInternet, onConfigureOnprem, onConfigureInternet }: ControlPanelProps) => {
    const {PlayButton, operationalStatus} = useAnimatePlayButton(knastInfo)

    return <div className="relative h-80">
        <KnastTypology x={0} y={22} showConnectivity={knastInfo?.state === Workstation_STATE_RUNNING} onpremHostsNumber={0} internetOpeningsNumber={3} onpremConnected={true} internetConnected={false}
            onConfigureInternet={onConfigureInternet}
            onConfigureOnprem={onConfigureOnprem}
            onConnectInternet={onConnectInternet}
            onConnectOnprem={onConnectOnprem}
            onDisconnectInternet={onDisconnectInternet}
            onDisconnectOnprem={onDisconnectOnprem}
        />
        <OperationalStatus operationalStatus={operationalStatus} />
        <PlayButton onButtonStart={onStartKnast} onButtonStop={onStopKnast} x={40} y={190}/>
        <IconButton x={200} y={190} ariaLabel="Settings" tooltip="Settings" onClick={onSettings}>
            <IconGear />
        </IconButton>
        <IconButton x={160} y={190} ariaLabel="Documentation" tooltip="Documentation" onClick={() => alert("Documentation")}>
            <IconOpenedBook />
        </IconButton>
    </div>
}