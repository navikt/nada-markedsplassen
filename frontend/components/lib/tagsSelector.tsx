import { Alert, Label } from '@navikt/ds-react'
import * as React from 'react'
import { useContext } from 'react'
import { ActionMeta, StylesConfig, ThemeConfig } from 'react-select'
import CreatableSelect from 'react-select/creatable'
import { UserState } from '../../lib/context'
import TagPill from './tagPill'
import { fetchKeywords, useFetchKeywords } from '../../lib/rest/keywords'

export interface TagsSelectorProps {
  onAdd: (value: string) => void
  onDelete: (value: string) => void
  tags: string[]
}

interface TagOption {
  value: string
  label: string
}

const styles: StylesConfig<TagOption, true> = {
  control: (provided: object, state: { isFocused: any }) => ({
    ...provided,
    border: state.isFocused ? '3px solid #00347d' : '1px solid',
  }),
  multiValue: (styles: object, { data }) => ({
    ...styles,
    padding: '0 0 0 0 !important',
    margin: '0 0 0 0 !important',
    width: '0px',
  }),
  multiValueLabel: (styles: object, { data }) => ({
    ...styles,
    padding: '0 0 0 0 !important',
    margin: '0 0 0 0 !important',
    width: '0px',
  }),
  multiValueRemove: (styles: object, { data }) => ({
    ...styles,
    padding: '0 0 0 0 !important',
    margin: '0 0 0 0 !important',
    width: '0px',
  }),
}

const theme: ThemeConfig = (t) => ({
  ...t,
  colors: {
    ...t.colors,
    neutral30: '#0067c5',
    neutral20: '#707070',
    primary: '#00347d',
  },
  spacing: {
    ...t.spacing,
    baseUnit: 3,
    controlHeight: 50,
  },
})

const useBuildTagOptionsList = () => {
  const keywordsList = useFetchKeywords()

  const userData = useContext(UserState)
  let tagsMap = new Map(
    userData?.dataproducts
      ?.flatMap((it: any) => it.keywords)
      .map((it: any) => [it, 100000])
  )
    keywordsList?.keywordItems.forEach((it: any) =>
      tagsMap.set(it.keyword, it.count + (tagsMap.get(it.keyword) || 0))
    )
  return Array.from(tagsMap.entries())
    .sort((l: any, r: any) => r[1] - l[1])
    .map((it) => it[0])
}

export const TagsSelector = ({ onAdd, onDelete, tags }: TagsSelectorProps) => {
  let tagOptions = useBuildTagOptionsList()
  const userData = useContext(UserState)
  const teamNames = userData?.allGoogleGroups?.map(
    (it: any) => it.name
  )

  const tagsLikeTeamName = tags.filter((it) =>
    teamNames?.find((tn: any) => tn.toLocaleLowerCase() === it.toLocaleLowerCase())
  )
  const addNewTag = (text: string) => {
    //if there is a similar tag in the list, use it as the text
    const newTag = (
      tagOptions.find(
        (t: any) => t && t.toLocaleLowerCase() === text.toLocaleLowerCase()
      ) as string ?? text
    ).trim()
    onAdd(newTag)
  }

  const onChangeInput = (_: any, actionMeta: ActionMeta<TagOption>) => {
    switch (actionMeta.action) {
      case 'pop-value':
      case 'remove-value':
        if (actionMeta.removedValue) {
          onDelete(actionMeta.removedValue.value)
        }
        break
      case 'select-option':
      case 'create-option':
        actionMeta.option && addNewTag(actionMeta.option.value)
        break
      default:
        console.log(`Unsupported action ${actionMeta.action}`)
    }
  }

  return (
    <div>
      <Label
        htmlFor={'0'}
        size={'medium'}
        className={'navds-text-field__label navds-label'}
      >
        Nøkkelord
      </Label>
      <br />
      <span className="italic text-[#555]">
        Team og PO blir automatisk lagt til og trenger ikke å oppgis som eget
        nøkkelord
      </span>
      {!!tagsLikeTeamName.length && (
        <Alert variant="info" size="small">
          {tagsLikeTeamName.map(
            (t, i) =>
              `${
                i === 0 ? '' : i === tagsLikeTeamName.length - 1 ? ' og ' : ', '
              }"${t}"`
          )}{' '}
          virker som {tagsLikeTeamName.length == 1 && 'et '}teamnavn og bør ikke
          oppgis som nøkkelord siden vi automatisk legger til team og
          PO-tilknytning.
        </Alert>
      )}
      <div className="flex flex-row gap-1 flex-wrap w-full mt-1 mb-1">
        {tags &&
          tags.map((k, i) => {
          
            return (
              <TagPill
                key={i}
                keyword={k}
                onClick={() => onDelete(k)}
                remove={true}
                horizontal={true}
              >
                {k}
              </TagPill>
            )
          })}
      </div>
      <CreatableSelect
        isMulti
        options={tagOptions.map((it) => ({
          value: it as string,
          label: it as string,
        }))}
        styles={styles}
        theme={theme}
        closeMenuOnSelect={false}
        placeholder="Legg til nøkkelord"
        onChange={onChangeInput}
        value={tags.map((it) => ({
          value: it,
          label: it,
        }))}
        isClearable={false}
      />
    </div>
  )
}
export default TagsSelector
