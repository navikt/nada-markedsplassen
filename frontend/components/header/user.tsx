import React, { useContext } from 'react'
import { useRouter } from 'next/router'
import { UserState } from '../../lib/context'
import { Dropdown, Header } from '@navikt/ds-react-internal'
import { Hamburger, People } from '@navikt/ds-icons'
import { Button } from '@navikt/ds-react'

export const backendHost = () => {
  return process.env.NODE_ENV === 'development' ? 'http://localhost:8080' : ''
}

const containsGroup = (groups: any[], groupEmail: string) => {
  for (let i = 0; i < groups.length; i++) {
    if (groups[i].email === groupEmail) return true
  }

  return false
}

export default function User() {
  const userData = useContext(UserState)
  const userOfNada = userData?.googleGroups.find((gr: any) => gr.name === 'nada')

  const router = useRouter()
  return userData ? (
    <div className="flex flex-row min-w-fit">
      <Dropdown>
        <Header.Button
          as={Dropdown.Toggle}
          className="border-transparent w-[48px] flex justify-center"
        >
          <Hamburger />
        </Header.Button>
        <Dropdown.Menu>
          <Dropdown.Menu.GroupedList>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={async () => await router.push('/dataproduct/new')}
            >
              Legg til nytt dataprodukt
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className="text-base flex gap-1 items-center"
              onClick={async () =>
                await router.push('/stories/new')
              }
            >
              Legg til ny datafortelling
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className="text-base flex gap-1 items-center"
              onClick={async () =>
                await router.push('/insightProduct/new')
              }
            >
              Legg til nytt innsiktsprodukt
            </Dropdown.Menu.GroupedList.Item>
            {<Dropdown.Menu.GroupedList.Item
              className="text-base flex gap-1 items-center"
              onClick={async () =>
                await router.push('/dataProc/joinableView/new')
              }
            >
              Koble sammen pseudonymiserte tabeller
            </Dropdown.Menu.GroupedList.Item>
            }
          </Dropdown.Menu.GroupedList>
          <Dropdown.Menu.Divider />
          <Dropdown.Menu.GroupedList>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/products' })
              }}
            >
              Mine produkter
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/stories' })
              }}
            >
              Mine fortellinger
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/insightProducts' })
              }}
            >
              Mine innsiktsprodukter
            </Dropdown.Menu.GroupedList.Item>

            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/requestsForGroup' })
              }}
            >
              Tilgangssøknader til meg
            </Dropdown.Menu.GroupedList.Item>

            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/requests' })
              }}
            >
              Mine tilgangssøknader
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/access' })
              }}
            >
              Mine tilganger
            </Dropdown.Menu.GroupedList.Item>
            <Dropdown.Menu.GroupedList.Item
              className={'text-base'}
              onClick={() => {
                router.push({ pathname: '/user/tokens' })
              }}
            >
              Mine team tokens
            </Dropdown.Menu.GroupedList.Item>
            {containsGroup(userData.googleGroups, "nada@nav.no") &&
              <Dropdown.Menu.GroupedList.Item
                className={'text-base'}
                onClick={() => {
                  router.push({ pathname: '/user/workstation' })
                }}
              >
                Min Knast
              </Dropdown.Menu.GroupedList.Item>
            }

            {userOfNada && <Dropdown.Menu.Divider />}
            {userOfNada && (
              <Dropdown.Menu.GroupedList.Item
                className={'text-base'}
                onClick={() => {
                  router.push({ pathname: '/admin/tags' })
                }}
              >
                Tags mantainence
              </Dropdown.Menu.GroupedList.Item>
            )}
          </Dropdown.Menu.GroupedList>
        </Dropdown.Menu>
      </Dropdown>

      <Dropdown>
        <Header.Button
          className="whitespace-nowrap hidden md:block text-base"
          as={Dropdown.Toggle}
        >
          <div className='flex flex-row'>
            <People className="h-[21px] w-[21px]" />
            {userData.name}
          </div>
        </Header.Button>
        <Dropdown.Menu>
          <Dropdown.Menu.GroupedList>
            <Dropdown.Menu.GroupedList.Item
              as="a"
              className={'text-base'}
              href={`${backendHost()}/api/logout`}
            >
              Logg ut
            </Dropdown.Menu.GroupedList.Item>
          </Dropdown.Menu.GroupedList>
        </Dropdown.Menu>
      </Dropdown>
    </div>
  ) : (
    <div className="flex flex-row min-w-fit">
      <Header.Button
        className={'h-full text-base'}
        onClick={async () =>
          await router.push(
            `${backendHost()}/api/login?redirect_uri=${encodeURIComponent(
              router.asPath
            )}`
          )
        }
        key="logg-inn"
      >
        Logg inn
      </Header.Button>
    </div>
  )
}
