import { Link } from "@navikt/ds-react"
import { ColorAuxText, ColorDefaultTextInvert, ColorDisabled, ColorInfoText, ColorSuccessful } from "../designTokens"
import { ProgressBar } from "./progressBar"
import { ExternalLinkIcon } from "@navikt/aksel-icons"
import { buildUrl } from "../../../lib/rest/apiUrl"
import { OpenKnastLink } from "./openKnastLink"

interface KnastDisplayProps {
    knastInfo: any
    operationalStatus: string
}

const LocalDev = () => {
    return <div>
        <Link>Lokal dev?</Link>
    </div>
}

export const KnastDisplay = ({ knastInfo, operationalStatus }: KnastDisplayProps) => {
    const bannerX = 22
    const bannerY = 40
    const bannerWidth = 220
    const bannerHeight = 30
    return <div className="relative">
        {operationalStatus === "started" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorSuccessful,
            color: ColorDefaultTextInvert,
        }}>
            Kjører
        </div>}
        {operationalStatus === "stopped" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorDisabled,
            color: ColorDefaultTextInvert,
        }}>
            Stoppet
        </div>}
        {operationalStatus === "restarting" && <ProgressBar x={bannerX} y={bannerY} height={bannerHeight} width={bannerWidth} color={ColorSuccessful} />}
        {operationalStatus === "restarting" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            color: ColorDefaultTextInvert,
        }}>
            Omstarter
        </div>}
        {operationalStatus === "starting" && <ProgressBar x={bannerX} y={bannerY} height={bannerHeight} width={bannerWidth} color={ColorSuccessful} />}
        {operationalStatus === "starting" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            color: ColorDefaultTextInvert,
        }}>
            Starter
        </div>}

        {operationalStatus === "stopping" && <ProgressBar x={bannerX} y={bannerY} height={bannerHeight} width={bannerWidth} color={ColorDisabled} />}
        {operationalStatus === "stopping" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            color: ColorDefaultTextInvert,
        }}>
            Stopper
        </div>}
        {operationalStatus === "started" ?
            <div className="absolute text-center" style={{
                left: 30,
                top: 100,
                width: 200,
                height: 20,
                color: ColorInfoText
            }}>
                <OpenKnastLink knastInfo={knastInfo} />
            </div>
            : <div>
                <div className="absolute text-center" style={{
                    left: 30,
                    top: 100,
                    width: 200,
                    height: 20,
                    color: ColorAuxText,
                }}>{knastInfo?.imageTitle}</div>
            </div>
        }

        {operationalStatus === "starting" && <div className="absolute flex flex-row text-sm justify-center items-center" style={{
            left: 30,
            top: 130,
            width: 210,
            color: ColorAuxText,
        }}> Det kan ta opptil 5 minutter...</div>
        }

    </div>
}