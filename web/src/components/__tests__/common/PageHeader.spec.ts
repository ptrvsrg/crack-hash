import { describe, it, expect } from 'vitest'

import { mount } from '@vue/test-utils'
import PageHeader from '@/components/header/PageHeader.vue'

describe('PageHeader', () => {
  it('renders properly', () => {
    const wrapper = mount(PageHeader)

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.isVisible()).toBe(true)
    expect(wrapper.text()).toContain('Crack HASH')
  })
})
