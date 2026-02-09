import React from "react";
import { IconArc, IconCircle, IconStartKnast, IconStopKnast } from "./knastIcons";
import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../../lib/rest/generatedDto";
import { Popover, Tooltip } from "@navikt/ds-react";
import { last } from "lodash";
import { ArrowsCirclepathIcon } from "@navikt/aksel-icons";
import { ColorInfoText } from "../designTokens";

type AnimateButtonAppearance = "started" | "stopped" | "stopping" | "starting" | "restarting"

interface AnimatePlayButtonProps {
    appearance?: AnimateButtonAppearance
    x: number | string
    y: number | string
    onButtonStart: () => void
    onButtonStop: () => void
    className?: string
}

const AnimatePlayButton = ({ appearance, x, y, onButtonStart, onButtonStop, className }: AnimatePlayButtonProps) => {
    const [hover, setHover] = React.useState(false)

    const getMainIconState = (appearance: AnimateButtonAppearance) =>
        (appearance === "starting" || appearance === "stopping" || appearance === "restarting") ? "grayed" : hover ? "hover" : "normal"

    const getDecorativeIconState = (appearance: AnimateButtonAppearance) =>
        (appearance === "starting" || appearance === "stopping" || appearance === "restarting") ? "normal" : "invisible"

    const bubbleStyle = {
        top: "-60px",
        left: "-40px",
        background: "#fff",
        border: "2px solid #ccc",
        padding: "2px",
        borderRadius: "8px",
        width: "110px"
    };

    const triOuter = {
        bottom: "-12px",
        left: "20px",
        width: 0,
        height: 0,
        borderLeft: "10px solid transparent",
        borderRight: "10px solid transparent",
        borderTop: "12px solid #ccc"
    };

    const triInner = {
        bottom: "-10px",
        left: "21px",
        width: 0,
        height: 0,
        borderLeft: "9px solid transparent",
        borderRight: "9px solid transparent",
        borderTop: "10px solid #fff"
    };

    return <div onClick={appearance === "stopped" ? onButtonStart : appearance === "started" ? onButtonStop : undefined}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        className={`absolute cursor-pointer ${className}`}
        style={{ left: x, top: y }}
    >
        <div className="absolute text-sm" style={bubbleStyle} hidden={appearance !== "stopped"}>
            klikk for å starte
            <span className="absolute" style={triOuter} />
            <span className="absolute" style={triInner} />
        </div>
        <Tooltip content={appearance === "stopped" ? "Start knast" : appearance === "started" ? "Stopp knast" : ""} hidden={appearance !== "started" && appearance !== "stopped"}>
            <div>
                <IconCircle state={appearance === "starting" || appearance === "stopping" || appearance === "restarting" ? "grayed" : getMainIconState(appearance ?? "stopped")}
                    className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
                <IconArc state={getDecorativeIconState(appearance ?? "stopped")} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 animate-spin" />
                {(appearance === "stopped" || appearance === "starting" || appearance === "restarting") &&
                    <IconStartKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
                {(appearance === "started" || appearance === "stopping") &&
                    <IconStopKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
            </div>
        </Tooltip>
    </div>
}

interface AnimateRestartButtonProps {
    appearance?: AnimateButtonAppearance
    x: number | string
    y: number | string
    onButtonRestart: () => void
    className?: string
}

const AnimateRestartButton = ({ appearance, x, y, onButtonRestart, className }: AnimateRestartButtonProps) => {
     return <div onClick={appearance === "started" ? onButtonRestart : undefined}
        className={`absolute cursor-pointer ${className}`}
        style={{ left: x, top: y }}
    >
        <Tooltip content={appearance === "started" ? "Omstart knast" : ""} hidden={appearance !== "started"}>
            <div>
                {appearance === "started" &&
                    <ArrowsCirclepathIcon color={ColorInfoText} fontSize={"2.2rem"}/>}
            </div>
        </Tooltip>
    </div>
}

export const useAnimatePlayButton = (knastInfo: WorkstationOutput | undefined) => {
    const [appearance, setAppearance] = React.useState<AnimateButtonAppearance>("stopped")
    const [lastRestarting, setLastRestarting] = React.useState<Date | undefined>(undefined)
    console.log("knastInfo state: " + knastInfo?.state)
    React.useEffect(() => {
        if(lastRestarting && lastRestarting > new Date(Date.now() - 10000)) {
            return
        }
        switch (knastInfo?.state) {
            case undefined:
            case Workstation_STATE_STOPPED:
                if (appearance !== "starting") {
                    setAppearance("stopped")
                }
                break
            case null:
                setAppearance("stopped")
                break
            case Workstation_STATE_STARTING:
                setAppearance("starting")
                break
            case Workstation_STATE_RUNNING:
                if (appearance !== "stopping") {
                    setAppearance("started")
                }
                break
            case Workstation_STATE_STOPPING:
                setAppearance("stopping")
                break
            default:
                setAppearance("stopped")
        }
    }, [knastInfo?.state, appearance])
    return {
        operationalStatus: appearance,
        PlayButton: (props: AnimatePlayButtonProps) => <AnimatePlayButton {...props}
            appearance={appearance}
            onButtonStart={() => {
                setAppearance("starting")
                props.onButtonStart()
            }}
            onButtonStop={() => {
                setAppearance("stopping")
                props.onButtonStop()
            }}
        ></AnimatePlayButton>,
        RestartButton: (props: AnimateRestartButtonProps) => <AnimateRestartButton {...props}
            appearance={appearance}
            onButtonRestart={() => {
                setAppearance("restarting")
                setLastRestarting(new Date())
                props.onButtonRestart()
            }}
        ></AnimateRestartButton>
    }
}