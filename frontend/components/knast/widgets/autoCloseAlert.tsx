import { Alert, AlertProps, Show } from "@navikt/ds-react";
import { useEffect, useState } from "react";

export const useAutoCloseAlert = (autoCloseMs: number) => {
    const [show, setShow] = useState<boolean>(false);

    useEffect(() => {
        if (show) {
            const timer = setTimeout(() => {
                setShow(false);
            }, autoCloseMs);
            return () => clearTimeout(timer);
        }
    }, [show, autoCloseMs]);

    return {
        showAlert: () => setShow(true),
        AutoHideAlert: (props: AlertProps) => (
            <Alert hidden={!show} {...props} />
        )
    };
}