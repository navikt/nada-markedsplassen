import React from "react";
import { DatasourcesForm } from "./DatasourcesForm";
import { InternetOpeningsForm } from "./internetOpeningsForm";
import { PythonConfigureForm } from "./pythonConfigureForm";
import { SettingsForm } from "./SettingsForm";


export interface ConfigureKnastFormProps {
    form: string
    knastData?: any
    knastOptions?: any
}

export const ConfigureKnastForm = ({ form, knastData, knastOptions }: ConfigureKnastFormProps) => {
    const [currentForm, setCurrentForm] = React.useState(form);

    const menuItemClassName = (itemName: string) =>
        itemName === currentForm
            ? "flex flex-row gap-2 font-bold border-l-4 pl-2"
            : "flex flex-row text-blue-600 cursor-pointer pl-2 border-l-4 border-transparent hover:border-b-1 hover:border-l-transparent hover:border-blue-600";
    return <div className="flex flex-row gap-6">
        <div className="flex flex-col gap-3 w-40">
            <div onClick={() => setCurrentForm("environment")} className={menuItemClassName("environment")}>
                Miljø
            </div>
            <div onClick={() => setCurrentForm("onprem")} className={menuItemClassName("onprem")}>
                Nav datakilder
            </div>
            <div onClick={() => setCurrentForm("internet")} className={menuItemClassName("internet")}>
                Internettåpninger
            </div>
            <div onClick={() => setCurrentForm("python")} className={menuItemClassName("python")}>
                Python
            </div>
        </div>
        <div>
            {currentForm === "environment" && <SettingsForm knastInfo={knastData} options={knastOptions?.data} />}
            {currentForm === "onprem" && <DatasourcesForm knastInfo={knastData} />}
            {currentForm === "internet" && <InternetOpeningsForm />}
            {currentForm === "python" && <PythonConfigureForm />}
        </div>
    </div>
}