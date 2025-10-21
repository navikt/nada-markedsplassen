import { Link } from "@navikt/ds-react"
import { buildUrl } from "../../../lib/rest/apiUrl"
import { ExternalLinkIcon } from "@navikt/aksel-icons"

export const OpenKnastLink = ({ knastInfo, caption: caption }: { knastInfo: any, caption?: string }) => {
    const handleOpenKnastInBrowser = () => {
        if (knastInfo?.host) {
            window.open(buildUrl('googleOauth2')('login')({ redirect: `https://${knastInfo.host}/` }), '_blank')
        }
    }

    return <Link onClick={handleOpenKnastInBrowser}> {caption??`Åpne ${knastInfo?.imageTitle}`}<ExternalLinkIcon /></Link>
}