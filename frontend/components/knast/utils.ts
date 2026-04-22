import { Workstation_STATE_RUNNING, Workstation_STATE_STARTING, Workstation_STATE_STOPPED, Workstation_STATE_STOPPING, WorkstationOutput } from "../../lib/rest/generatedDto"

export const getOperationalStatus = (ws: WorkstationOutput | undefined): string =>{
    switch(ws?.state) {
        case Workstation_STATE_RUNNING:
            return "Running"
        case Workstation_STATE_STARTING:
            return "Starting"
        case Workstation_STATE_STOPPING:
            return "Stopping"
        case Workstation_STATE_STOPPED:
            return "Stopped"
        default:
            return "Unknown"
    }
}

export const getKnastDailyCost = (knastInfo: any) => knastInfo?.machineTypeInfo?.hourlyCost ? `Kr ${(knastInfo.machineTypeInfo.hourlyCost * 24).toFixed(0)},-/døgn` : "Ukjent kostnad"
export const isValidUrl = (url: string) => {
    // Per Google Secure Web Proxy UrlList syntax:
    // https://cloud.google.com/secure-web-proxy/docs/url-list-syntax-reference
    // - No scheme, no query string (?...), no fragment (#...).
    // - Host: dot-separated labels; the leftmost label may start with a single '*' wildcard.
    // - TLD: 2+ alphanumeric/hyphen characters.
    // - Optional path starts with '/', allows RFC 3986 unreserved/sub-delims/pchar
    //   characters plus '*' wildcard, but never '?' or '#'.
    const urlPattern = /^((\*|\*?[a-zA-Z0-9-]+)\.)+[a-zA-Z0-9-]{2,}(\/[a-zA-Z0-9-._~:@!$&'()*+,;=/]*)?$/;
    return !url ? true : urlPattern.test(url);
}