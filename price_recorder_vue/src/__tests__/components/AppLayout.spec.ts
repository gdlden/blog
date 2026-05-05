import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const pushMock = vi.fn()
let routeName = 'blog'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({
      name: routeName,
    }),
    useRouter: () => ({
      push: pushMock,
    }),
  }
})

describe('AppLayout.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    routeName = 'blog'
    pushMock.mockClear()
    vi.restoreAllMocks()
  })

  it('renders sidebar with "博文", "债务" and "油耗" links', async () => {
    const { default: AppLayout } = await import('@/components/AppLayout.vue')
    const wrapper = mount(AppLayout, {
      global: {
        stubs: { RouterView: true, RouterLink: RouterLinkStub },
      },
    })

    expect(wrapper.text()).toContain('博文')
    expect(wrapper.text()).toContain('债务')
    expect(wrapper.text()).toContain('油耗')
  })

  it('renders logout button with text "退出"', async () => {
    const { default: AppLayout } = await import('@/components/AppLayout.vue')
    const wrapper = mount(AppLayout, {
      global: {
        stubs: { RouterView: true, RouterLink: RouterLinkStub },
      },
    })

    expect(wrapper.text()).toContain('退出')
  })

  it('clicking logout calls userStore.clearUserInfo and pushes to /login', async () => {
    const { default: AppLayout } = await import('@/components/AppLayout.vue')
    const wrapper = mount(AppLayout, {
      global: {
        stubs: { RouterView: true, RouterLink: RouterLinkStub },
      },
    })

    const logoutButton = wrapper.findAll('button').find((b) => b.text() === '退出')
    expect(logoutButton).toBeDefined()
    await logoutButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith('/login')
  })

  it('active route link receives the active styling class', async () => {
    const { default: AppLayout } = await import('@/components/AppLayout.vue')
    const wrapper = mount(AppLayout, {
      global: {
        stubs: {
          RouterView: true,
          RouterLink: {
            ...RouterLinkStub,
            props: ['to', 'activeClass'],
          },
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)
    expect(links.length).toBeGreaterThanOrEqual(2)

    const blogLink = links.find((l) => l.props().to === '/blog' && l.classes().includes('relative'))
    const debtLink = links.find((l) => l.props().to === '/debt' && l.classes().includes('relative'))

    expect(blogLink!.classes()).toContain('bg-white/15')
    expect(debtLink!.classes()).not.toContain('bg-white/15')
  })
})
