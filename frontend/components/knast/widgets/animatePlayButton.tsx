import React from "react";
import { IconArc, IconCircle, IconStartKnast, IconStopKnast } from "./knastIcons";
import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../../lib/rest/generatedDto";

type AnimateButtonAppearance = "Started" | "Stopped" | "Stopping" | "Starting"

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
        (appearance === "Starting" || appearance === "Stopping") ? "grayed" : hover ? "hover" : "normal"

    const getDecorativeIconState = (appearance: AnimateButtonAppearance) =>
        (appearance === "Starting" || appearance === "Stopping") ? "normal" : "invisible"

    return <div onClick={appearance === "Stopped" ? onButtonStart : appearance === "Started" ? onButtonStop : undefined}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        className={`absolute ${className}`}
        style={{ left: x, top: y }}
    >
        <IconCircle state={appearance === "Starting" || appearance === "Stopping" ? "grayed" : getMainIconState(appearance ?? "Stopped")}
            className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
        <IconArc state={getDecorativeIconState(appearance ?? "Stopped")} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 animate-spin" />
        {(appearance === "Stopped" || appearance === "Starting") &&
            <IconStartKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
        {(appearance === "Started" || appearance === "Stopping") &&
            <IconStopKnast state={getMainIconState(appearance)} className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />}
    </div>
}

export const useAnimatePlayButton = (knastInfo: WorkstationOutput | undefined) => {
    const [appearance, setAppearance] = React.useState<AnimateButtonAppearance>("Stopped")
    React.useEffect(() => {
        switch (knastInfo?.state) {
            case undefined:
            case Workstation_STATE_STOPPED:
                if (appearance !== "Starting") {
                    setAppearance("Stopped")
                }
                break
            case null:
                setAppearance("Stopped")
                break
            case Workstation_STATE_STARTING:
                setAppearance("Starting")
                break
            case Workstation_STATE_RUNNING:
                if (appearance !== "Stopping") {
                    setAppearance("Started")
                }
                break
            case Workstation_STATE_STOPPING:
                setAppearance("Stopping")
                break
            default:
                setAppearance("Stopped")
        }
    }, [knastInfo?.state])
    return {
        operationalStatus: appearance,
        PlayButton: (props: AnimatePlayButtonProps) => <AnimatePlayButton {...props} 
        appearance={appearance}
        onButtonStart={()=>{
            setAppearance("Starting")
            props.onButtonStart()
        }}
        onButtonStop={()=>{
            setAppearance("Stopping")
            props.onButtonStop()
        }}
        ></AnimatePlayButton>
    }
}