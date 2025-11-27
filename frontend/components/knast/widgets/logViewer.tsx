import { CopyButton, Loader, Table } from "@navikt/ds-react";
import { useUpdateUrlAllowList, useWorkstationLogs } from "../queries";
import { formatDistanceToNow } from "date-fns";
import React from "react";
import { ColorFailed, ColorInfoText } from "../designTokens";
import { UrlTextDisplay } from "./urlItem";
import { LogEntry, WorkstationLogs } from "../../../lib/rest/generatedDto";

interface BlockedURLLogProps {
    logs: WorkstationLogs;
    entry: LogEntry;
}

const translateTime = (timeEng: string) =>{
    return timeEng.replace("seconds", "sekunder")
        .replace("second", "sekund")
        .replace("minutes", "minutter")
        .replace("minute", "minutt")
        .replace("hours", "timer")
        .replace("hour", "time")
        .replace("days", "dager")
        .replace("day", "dag")
        .replace("weeks", "uker")
        .replace("week", "uke")
        .replace("months", "måneder")
        .replace("month", "måned")
        .replace("years", "år")
        .replace("year", "år")
        .replace("ago", "siden")
        .replace("about ", "ca.");
}

const BlockedURLLog = ({ logs, entry }: BlockedURLLogProps) => {
    return <div className="grid grid-cols-[20%_80%] border-b border-gray-300 p-2">
        <div className="text-sm">
            {isNaN(new Date(entry.Timestamp).getTime())
                ? 'Ugyldig dato'
                : translateTime(formatDistanceToNow(new Date(entry.Timestamp), { addSuffix: true }))
            }
        </div>
        <div>
            <div className="flex flex-row gap-2 justify-between">
                <div title={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`} className="flex flex-row gap-2">
                    <div style={{
                        color: ColorFailed
                    }}>URL blokkert</div>
                    <div className="flex flex-row items-start">
                        <CopyButton
                            title="Kopier URL"
                            color={ColorInfoText}
                            copyText={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`}
                            aria-text="Kopier URL"
                            size="xsmall"
                        />
                        <UrlTextDisplay url={`${entry.HTTPRequest?.URL.Host}${entry.HTTPRequest?.URL.Path}`} lengthLimitHead={20} lengthLimitTail={10} />
                    </div>
                </div>
            </div>
        </div>
    </div>;
}

export const LogViewer = () => {
    const logs = useWorkstationLogs()
    const aggregatedLogs = logs.data?.proxyDeniedHostPaths ?? [];
    console.log(logs.data)
    return (
        <div className="w-160 h-140 border-blue-100 border rounded p-4">

            <div className="mt-2 mb-4">
                <h4>Hendelseslogg</h4>
            </div>
            <div className="bg-gray-100">
                <div className="grid grid-cols-[20%_80%] border-b border-gray-300 p-2">
                    <div>
                        Tidspunkt
                    </div>
                    <div>
                        Meldling
                    </div>
                </div>

            </div>
            {logs.isLoading && <Loader className="mt-6" size="small" title="Laster.." />}
            {aggregatedLogs.length > 0 ? <div className="overflow-y-auto h-110">
                {aggregatedLogs.map((url: any, i: number) => (<div key={i}>
                    <BlockedURLLog key={i + url.HTTPRequest.URL.Host + url.Timestamp} logs={logs.data!} entry={url} />
                </div>))}
            </div>
                : <div className="h-full text-sm flex items-center justify-center text-center">Ingen logger funnet</div>}
        </div >
    );
}