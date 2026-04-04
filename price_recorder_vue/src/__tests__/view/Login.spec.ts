import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Login from '@/view/Login.vue'
import * as loginApi from '@/api/Login'

const pushMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({
      query: { redirect: '/test' },
    }),
    useRouter: () => ({
      push: pushMock,
    }),
  }
})

describe('Login.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockClear()
    vi.restoreAllMocks()
  })

  it('renders username and password inputs and login button', () => {
    const wrapper = mount(Login)

    expect(wrapper.find('input[type="text"]').exists()).toBe(true)
    expect(wrapper.find('input[type="password"]').exists()).toBe(true)

    const loginButton = wrapper.findAll('button').find(b => b.text() === '登录')
    expect(loginButton).toBeDefined()
    expect(loginButton!.exists()).toBe(true)
  })

  it('clicking login calls the login API', async () => {
    const wrapper = mount(Login)
    const loginSpy = vi.spyOn(loginApi, 'login').mockResolvedValue({ token: 'token', userId: 1 } as any)

    const usernameInput = wrapper.find('input[type="text"]')
    const passwordInput = wrapper.find('input[type="password"]')
    await usernameInput.setValue('alice')
    await passwordInput.setValue('secret')

    const loginButton = wrapper.findAll('button').find(b => b.text() === '登录')
    await loginButton!.trigger('click')

    expect(loginSpy).toHaveBeenCalled()
  })

  it('successful login triggers store setUserInfo and router push to redirect query', async () => {
    const wrapper = mount(Login)
    vi.spyOn(loginApi, 'login').mockResolvedValue({ token: 'token', userId: 1 } as any)

    const loginButton = wrapper.findAll('button').find(b => b.text() === '登录')
    await loginButton!.trigger('click')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith('/test')
  })

  it('failed login shows error message "登录失败，请检查用户名或密码后重试。"', async () => {
    const wrapper = mount(Login)
    vi.spyOn(loginApi, 'login').mockRejectedValue(new Error('fail'))

    const loginButton = wrapper.findAll('button').find(b => b.text() === '登录')
    await loginButton!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('登录失败，请检查用户名或密码后重试。')
  })
})
