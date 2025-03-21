import dynamic from 'next/dynamic'
import { FieldValues, useController, UseControllerProps } from 'react-hook-form'
import Link from 'next/link'
import { Loader } from '@navikt/ds-react'

const MDEditor = dynamic(() => import('@uiw/react-md-editor'), {
  ssr: false,
  loading: () => (
    <div className="h-48 mt-2 w-full flex items-center justify-center border border-border-strong rounded-sm">
      <Loader size="2xlarge" />
    </div>
  ),
})

type DescriptionEditorProps<T extends FieldValues> = { label: string } & UseControllerProps<T>

export const DescriptionEditor = <FV extends Record<string, any>>({
  name,
  control,
  label,
}: DescriptionEditorProps<FV>) => {
  const {
    field: { ...inputProps },
  } = useController({
    name,
    control,
  })

  return (
    <div data-color-mode="light">
      <div className="flex flex-col">
        <label
          style={{ display: 'inline' }}
          htmlFor={name}
          className={'navds-text-field__label navds-label'}
        >
          {label}
        </label>
        <span className="italic text-[#555]">
          Formattering i Markdown,{' '}
          <Link href={'https://guides.github.com/features/mastering-markdown/'}>
            se innføring
          </Link>
        </span>
      </div>
      <MDEditor
        className="w-full navds-body-short navds-body-medium mt-2 border border-border-strong shadow-none"
        {...inputProps}
        preview="edit"
        style={{}}
      />
    </div>
  )
}

export default DescriptionEditor
