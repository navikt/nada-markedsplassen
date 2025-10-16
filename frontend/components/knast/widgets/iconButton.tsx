import React from "react";

interface IconButtonProps {
    x: number | string;
    y: number | string;
    className?: string;
    tooltip?: string;
    ariaLabel?: string;
    onClick?: () => void;
    children: React.ReactNode;
}

export const IconButton = ({ x, y, className, tooltip, ariaLabel, onClick, children }: IconButtonProps) => {
    const [hover, setHover] = React.useState(false)
    const childWithHover = React.cloneElement(children as React.ReactElement, { state: hover ? "hover" : "normal"})
    return <div className={`absolute -translate-x-1/2 -translate-y-1/2 ${className}`} style={{ left: x, top: y }} onClick={onClick} aria-label={ariaLabel} title={tooltip}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
    >
        {childWithHover}
    </div>
}