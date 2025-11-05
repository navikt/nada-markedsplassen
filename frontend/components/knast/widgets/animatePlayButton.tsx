import React from "react";
import { IconArc, IconCircle, IconStartKnast, IconStopKnast } from "./knastIcons";
import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../../lib/rest/generatedDto";

type AnimateButtonAppearance = "started" | "stopped" | "stopping" | "starting"

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
        (appearance === "starting" || appearance === "stopping") ? "grayed" : hover ? "hover" : "normal"

    const getDecorativeIconState = (appearance: AnimateButtonAppearance) =>
        (appearance === "starting" || appearance === "stopping") ? "normal" : "invisible"

    return <div onClick={appearance === "stopped" ? onButtonStart : appearance === "started" ? onButtonStop : undefined}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        className={`absolute ${className}`}
        style={{ left: x, top: y }}
    >
        <IconCircle state={appearance === "starting" || appearance === "stopping" ? "grayed" : getMainIconState(appearance ?? "stopped")}
            className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
        <IconArc state={getDecorativeIconState(appearance ?? "stopped")} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 animate-spin" />
        {(appearance === "stopped" || appearance === "starting") &&
            <IconStartKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
        {(appearance === "started" || appearance === "stopping") &&
            <IconStopKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
    </div>
}

export const useAnimatePlayButton = (knastInfo: WorkstationOutput | undefined) => {
    const [appearance, setAppearance] = React.useState<AnimateButtonAppearance>("stopped")
    React.useEffect(() => {
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
        onButtonStart={()=>{
            setAppearance("starting")
            props.onButtonStart()
        }}
        onButtonStop={()=>{
            setAppearance("stopping")
            props.onButtonStop()
        }}
        ></AnimatePlayButton>
    }
}