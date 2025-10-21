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
    const bannerWidth = 200
    const bannerHeight = 30
    return <div className="relative">
        {operationalStatus === "Started" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorSuccessful,
            color: ColorDefaultTextInvert,
        }}>
            Kjører
        </div>}
        {operationalStatus === "Stopped" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            backgroundColor: ColorDisabled,
            color: ColorDefaultTextInvert,
        }}>
            Stoppet
        </div>}
        {operationalStatus === "Starting" && <ProgressBar x={bannerX} y={bannerY} height={bannerHeight} width={bannerWidth} color={ColorSuccessful} />}
        {operationalStatus === "Starting" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            color: ColorDefaultTextInvert,
        }}>
            Starter
        </div>}
        {operationalStatus === "Stopping" && <ProgressBar x={bannerX} y={bannerY} height={bannerHeight} width={bannerWidth} color={ColorDisabled} />}
        {operationalStatus === "Stopping" && <div className="absolute text-center flex flex-col justify-center" style={{
            left: bannerX,
            top: bannerY,
            width: bannerWidth,
            height: bannerHeight,
            color: ColorDefaultTextInvert,
        }}>
            Stopper
        </div>}
        {operationalStatus === "Started" ?
            <div className="absolute text-center" style={{
                left: 30,
                top: 90,
                width: 180,
                height: 20,
                color: ColorInfoText
            }}>
                <OpenKnastLink knastInfo={knastInfo} />
            </div>
            : <div>
                <div className="absolute text-center" style={{
                    left: 30,
                    top: 90,
                    width: 180,
                    height: 20,
                    color: ColorDisabled,
                }}>{knastInfo?.imageTitle}</div>
            </div>
        }

        {operationalStatus === "Started" &&
            <div className="absolute text-center" style={{
                left: 30,
                top: 130,
                width: 180,
                color: ColorInfoText,
            }}><LocalDev /></div>
        }

        {operationalStatus === "Starting" && <div className="absolute text-sm" style={{
            left: 30,
            top: 120,
            width: 180,
            color: ColorAuxText,
        }}> Det kan ta opptil 5 minutter, vennligst vent</div>
        }

    </div>
}