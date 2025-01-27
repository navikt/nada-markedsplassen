import { MagnifyingGlassIcon } from '@navikt/aksel-icons'
import React, { FormEvent, useState } from 'react'

interface SearchBoxProps {
  onSearch: (query: string) => void
}

const SearchBox = ({ onSearch }: SearchBoxProps) => {
  const [value, setValue] = useState<string>('')

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    onSearch(value)
  }

  return (
    <div role="navigation" className="flex justify-center">
      <form
        className="flex items-center border-2 border-blue-300 rounded-lg w-10/12 px-3 py-1 min-w-[350px] max-w-[500px]"
        onSubmit={onSubmit}
      >
        <input
          className="w-full border-0 p-0 focus-visible:outline-none"
          type="search"
          aria-label="Søkefelt"
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />

        <button id="some-id" aria-label="Søk">
          <MagnifyingGlassIcon className="w-7 h-7" />
        </button>
      </form>
    </div>
  )
}
export default SearchBox
