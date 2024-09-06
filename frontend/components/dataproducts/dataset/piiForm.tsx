import { AnnotateColumnListener, ColumnType, PIITagType, PseudoColumnListener } from "./useColumnTags"
import { Control, Controller, FormState, UseFormGetValues, UseFormRegister, UseFormWatch } from "react-hook-form"
import { Alert, Radio, RadioGroup, Textarea, Switch } from "@navikt/ds-react"
import { Personopplysninger, PseudonymiseringsText } from "./helptext"
import AnnotateDatasetTable from "./annotateDatasetTable"
import { useState } from "react"
import { FormValues } from "./newDatasetForm"

interface PiiFormProps {
    loading: boolean
    error: any
    columns: ColumnType[] | undefined
    tags: Map<string, PIITagType> | undefined
    pseudoColumns: Map<string, boolean>
    control: Control<FormValues, any>
    getValues: UseFormGetValues<FormValues>
    register: UseFormRegister<FormValues>
    formState: FormState<FormValues>
    watch: UseFormWatch<FormValues>
    annotateColumn: AnnotateColumnListener
    pseudoynimiseColumn: PseudoColumnListener
}

export const PiiForm = ({
    loading,
    error,
    columns,
    tags,
    pseudoColumns,
    control,
    getValues,
    register,
    formState,
    watch,
    annotateColumn,
    pseudoynimiseColumn,
}: PiiFormProps) => {

    var showAnnotateDatasetTable = watch("pii") === "sensitive"
    var [createPseudoynimizedView, setcreatePseudoynimizedView] = useState(false)

    return <div>
        <Controller
            name="pii"
            control={control}
            render={({ field }) => (
                <RadioGroup
                    {...field}
                    legend={
                        <p className="flex gap-2 items-center">
                            Inneholder datasettet personopplysninger?{' '}
                            <Personopplysninger />
                        </p>
                    }
                >
                    <Radio value={'sensitive'}>
                        Ja, inneholder personopplysninger
                    </Radio>
                    {showAnnotateDatasetTable &&
                        <div className="flex items-center gap-x-1">
                            <Switch onChange={e=>{
                                setcreatePseudoynimizedView(e.target.checked)
                            }}>Jeg ønsker å pseudonymisere tabellen </Switch>
                            <PseudonymiseringsText />
                        </div>
                    }
                    {createPseudoynimizedView && <Alert variant="info">Velg kolonner du ønsker å pseudonymisere</Alert>}
                    {showAnnotateDatasetTable && (
                        <AnnotateDatasetTable
                            loading={loading}
                            error={error}
                            columns={columns}
                            tags={tags}
                            pseudoColumns={pseudoColumns}
                            selectPseudoColumn={createPseudoynimizedView ? pseudoynimiseColumn:undefined}
                            annotateColumn={annotateColumn}
                        />
                    )}
                    <Radio value={'anonymised'}>
                        Det er benyttet metoder for å anonymisere personopplysningene
                    </Radio>
                    <Textarea
                        placeholder="Beskriv kort hvordan opplysningene er anonymisert"
                        label="Metodebeskrivelse"
                        aria-hidden={getValues('pii') !== 'anonymised'}
                        className={getValues('pii') !== 'anonymised' ? 'hidden' : ''}
                        error={formState.errors?.anonymisation_description?.message?.toString()}
                        {...register('anonymisation_description')}
                    />
                    <Radio value={'none'}>
                        Nei, inneholder ikke personopplysninger
                    </Radio>
                </RadioGroup>
            )}
        />
    </div>
} 