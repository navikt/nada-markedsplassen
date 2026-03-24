import { Tag } from '@navikt/ds-react'
import React from 'react'
import TagRemoveIcon from './icons/tagRemoveIcon'

export const KeywordBox = ({ children }: { children: React.ReactNode }) => (
  <div className="flex flex-row gap-1 flex-wrap justify-end">{children}</div>
)

interface keywordPillProps {
  keyword: string
  compact?: boolean
  children?: React.ReactNode
  onClick?: () => void
  remove?: boolean
  lineThrough?: boolean
}

export const TagPill = ({
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
        className={`text-ax-text-neutral flex items-center bg-ax-bg-neutral-soft border-ax-border-neutral
      ${onClick && 'cursor-pointer'}
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
