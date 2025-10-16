import { Button } from "@navikt/ds-react";

interface InternetOpeningsFormProps {
    onSave: () => void;
    onCancel: () => void;
}

export const InternetOpeningsForm = ({ onSave, onCancel }: InternetOpeningsFormProps) => {
    return <div className="max-w-[35rem] border-blue-100 border rounded p-4">
        <p>Here you can configure internet access settings for your workstation.</p>
        <p>This feature is coming soon!</p>
        <Button variant="primary" onClick={onSave}>Save</Button>
        <Button variant="secondary" className="ml-6" onClick={onCancel}>Cancel</Button>
    </div>
}