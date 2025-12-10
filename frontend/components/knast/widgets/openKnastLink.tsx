import { Button, Link, Popover, Tooltip } from "@navikt/ds-react"
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

    return <div className = "flex flex-row justify-center items-center">
        <Popover
            className="w-60"
            open={showNaisdeviceInfo}
            onClose={() => setShowNaisdeviceInfo(false)}
            anchorEl={linkRef.current}
        >
            <NaisdevicePopoverContent />
        </Popover>
        <Button size="small" className="flex flex-row gap-2" 
        style={{
            cursor: 'pointer'
        }}
        onMouseEnter={() => setShowNaisdeviceInfo(true)}
            onMouseLeave={() => setShowNaisdeviceInfo(false)}
            onClick={handleOpenKnastInBrowser}> {caption ?? `Åpne ${knastInfo?.imageTitle}`}</Button>
    </div>
}