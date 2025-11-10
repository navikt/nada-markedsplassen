import { MenuHamburgerIcon } from '@navikt/aksel-icons'
import { Dropdown, InternalHeader, Loader } from '@navikt/ds-react'
import { useRouter } from 'next/router'
import { useContext } from 'react'
import { UserState } from '../../lib/context'

export default function User() {
  const userData = useContext(UserState)
  const userOfNada = userData?.googleGroups?.find((gr: any) => gr.name === 'nada')

  const router = useRouter()
  return userData ? (
    <div className="flex flex-row min-w-fit">
      <style jsx>{`
        .blinking {
          animation: blink 2s infinite;
        }

        @keyframes blink {
          0% { color: #ff8080; }
          50% { color: #ffff80; }
          100% { color: #ff8080; }
        }
      `}
      </style>
      {userData.isKnastUser && <InternalHeader.Button
        className="border-transparent w-[8rem] flex justify-center"
        onClick={async () => await router.push('/user/workstation')}
      >
        KNAST
      </InternalHeader.Button>}
      <Dropdown>
        <InternalHeader.Button
          as={Dropdown.Toggle}
          className="border-transparent w-[48px] flex justify-center"
        >
          <MenuHamburgerIcon />
        </InternalHeader.Button>
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
            <Dropdown.Menu.GroupedList.Item
              className="text-base flex gap-1 items-center"
              onClick={async () =>
                await router.push('/metabaseDashboard/new')
              }
            >
              Legg til nytt public Metabase dashboard
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
                router.push({ pathname: '/user/publicDashboards' })
              }}
            >
              Mine public Metabase dasboards
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
            {userData.isKnastUser &&
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
                Tags maintenance
              </Dropdown.Menu.GroupedList.Item>
            )}
            {userOfNada && (
              <Dropdown.Menu.GroupedList.Item
                className={'text-base'}
                onClick={() => {
                  router.push({ pathname: '/admin/knast' })
                }}
              >
                Knast maintenance
              </Dropdown.Menu.GroupedList.Item>
            )}
            {userOfNada && (
              <Dropdown.Menu.GroupedList.Item
                className={'text-base'}
                onClick={() => {
                  router.push({ pathname: '/river' })
                }}
              >
                River jobs maintenance
              </Dropdown.Menu.GroupedList.Item>
            )}
          </Dropdown.Menu.GroupedList>
        </Dropdown.Menu>
      </Dropdown>
      <InternalHeader.User
        name={userData.name}
      />
    </div>
  ) : (
    <div className="flex flex-row min-w-fit pr-3">
      <Loader variant='inverted' />
    </div>
  )

}
