import { Tag } from '@navikt/ds-react'
import React from 'react'
import TagRemoveIcon from './icons/tagRemoveIcon'

export const KeywordBox = ({ children }: { children: React.ReactNode }) => (
  <div className="flex flex-row gap-1 flex-wrap justify-end">{children}</div>
)

interface keywordPillProps {
  keyword: string
  horizontal?: boolean
  compact?: boolean
  children?: React.ReactNode
  onClick?: () => void
  remove?: boolean
  lineThrough?: boolean
}

export const TagPill = ({
  horizontal,
  lineThrough,
  children,
  onClick,
  remove,
}: keywordPillProps) => {
  return (
    <div className="flex algin-middle">
      <Tag
        variant="info"
        size="small"
        onClick={onClick}
        className={`text-text flex items-center bg-surface-subtle border-border-default
      ${onClick && 'cursor-pointer'}
      ${horizontal ? 'flex' : 'inline-block'}
      ${remove && 'hover:decoration-[3px] hover:line-through'}
      ${lineThrough && 'decoration-[1px] line-through'}`}
      >
        {children}
        {remove && <div className={`h-2rem pl-1 place-items-center `}>
          <TagRemoveIcon />
        </div>}
      </Tag>
    </div>
  )
}

export default TagPill
