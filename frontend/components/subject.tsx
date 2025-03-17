import React from 'react'

export const Subject = ({ children }: { children: React.ReactNode }) => {
  return <div className="mb-5 ml-[1px] text-base text-text-default">{children}</div>
}

export const SubjectHeader = ({
  centered,
  children,
}: {
  centered?: boolean
  children: React.ReactNode
}) => {
  return (
    <h2
      className={`${
        centered ? 'mx-auto ' : ''
      }pb-0 mt-0 mb-1 text-text-default font-medium! text-xs!`}
    >
      {children}
    </h2>
  )
}
