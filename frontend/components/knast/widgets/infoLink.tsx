import { Link, Popover } from "@navikt/ds-react";

export const InfoLink = ({ caption, content, className }: { caption: string, content: React.ReactNode, className?: string }) => {
    const [showInfo, setShowInfo] = React.useState(false);
    const linkRef = React.useRef(null);

    return (
        <div>
            <Popover
                className="w-60"
                open={showInfo}
                onClose={() => setShowInfo(false)}
                anchorEl={linkRef.current}
            >
                {content}
            </Popover>
            <Link ref={linkRef} href="#" onClick={() => setShowInfo(!showInfo)} className={className}>{caption}</Link>
        </div>
    );
}
