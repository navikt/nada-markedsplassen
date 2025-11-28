import { Link, Popover, Tooltip } from "@navikt/ds-react"
import { buildUrl, buildUrlAsClientWithoutProxy } from "../../../lib/rest/apiUrl"
import { ExternalLinkIcon, InformationSquareFillIcon } from "@navikt/aksel-icons"
import { NaisdeviceGreen } from "../../lib/icons/naisdeviceGreen"
import React from "react"

const NaisdevicePopoverContent = () => {
    return (
        <Popover.Content>
            <div className="flex flex-row items-end">
                <NaisdeviceGreen />

                Er naisdevice &nbsp;
                <p className="text-[#78c010]">grønn?</p>
            </div>
        </Popover.Content>
    )
}

export const OpenKnastLink = ({ knastInfo, caption: caption }: { knastInfo: any, caption?: string }) => {
    const [showNaisdeviceInfo, setShowNaisdeviceInfo] = React.useState(false);
    const linkRef = React.useRef<HTMLAnchorElement | null>(null);
    const handleOpenKnastInBrowser = () => {
        if (knastInfo?.host) {
            window.open(buildUrlAsClientWithoutProxy('googleOauth2')('login')({ redirect: `https://${knastInfo.host}/` }), '_blank')
        }
    }

    return <div>
        <Popover
            className="w-60"
            open={showNaisdeviceInfo}
            onClose={() => setShowNaisdeviceInfo(false)}
            anchorEl={linkRef.current}
        >
            <NaisdevicePopoverContent />
        </Popover>
        <Link 
        style={{
            cursor: 'pointer'
        }}
        onMouseEnter={() => setShowNaisdeviceInfo(true)}
            onMouseLeave={() => setShowNaisdeviceInfo(false)}
            onClick={handleOpenKnastInBrowser} ref={linkRef}> {caption ?? `Åpne ${knastInfo?.imageTitle}`}<ExternalLinkIcon /></Link>
    </div>
}