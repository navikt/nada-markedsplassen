import { ColorSuccessful } from "../designTokens";

// ProgressBar.tsx
export const ProgressBar = ({
    x= 0,
    y= 0,
    height = "100%",
    width = "100%",
    color = ColorSuccessful,
}: {
    x?: number | string;
    y?: number | string;
    height?: number | string;
    width?: number | string;
    color?: string;
}) => {

    const strip: React.CSSProperties = {
        backgroundImage:
            "repeating-linear-gradient(45deg, rgba(255,255,255,0.25) 0, rgba(255,255,255,0.25) 10px, rgba(255,255,255,0.1) 10px, rgba(255,255,255,0.1) 20px)",
        backgroundSize: "200px 30px",
    };

    return (
            <div
                className="absolute overflow-hidden"
                style={{ left: x, top: y, height, width }}
                role="progressbar"
                aria-label="progress"
                aria-valuemin={0}
                aria-valuemax={100}
            >
                {/* Moving fill (clipped by parent) */}
                <div
                    className={`absolute inset-y-0 left-0 animate-stripes`}
                    style={{
                        width: "200%",
                        ...strip,
                        backgroundColor: color,
                    }}
                />
            </div>
    );
};
