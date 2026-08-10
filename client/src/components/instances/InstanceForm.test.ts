import { describe, expect, it } from 'vitest'

import { mount } from '@vue/test-utils'

import InstanceForm from './InstanceForm.vue'

describe('InstanceForm', () => {
  it('submits valid data', async () => {
    const wrapper = mount(InstanceForm)

    await wrapper.get('input[name="name"]').setValue('my-db-2')

    await wrapper.get('input[name="instances"]').setValue('2')

    await wrapper.get('input[name="version"]').setValue('15')

    await wrapper.get('input[name="storage"]').setValue('7Gi')

    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('submit')).toEqual([
      [
        {
          name: 'my-db-2',
          instances: 2,
          version: 15,
          storage: '7Gi',
        },
      ],
    ])
  })

  // it('shows validation errors', async () => {
  //   const wrapper = mount(InstanceForm)

  //   await wrapper.get('form').trigger('submit')

  //   expect(wrapper.text()).toContain('Name is required')

  //   expect(wrapper.text()).toContain('At least 1 instance is required')

  //   expect(wrapper.emitted('submit')).toBeUndefined()
  // })
})
