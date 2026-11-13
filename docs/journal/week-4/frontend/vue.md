# Vue.js 3 — Practical Notes for an Advanced React Developer

> **Target:** React/Next.js developer moving to Vue 3
> **Approach:** Learn the Vue mental model first, then map it to React.
> **Recommended Vue style:** Composition API + `<script setup>` + Single-File Components (SFCs).

---

# 1. What is Vue?

**Vue** is a progressive JavaScript framework for building user interfaces.

It provides:

* Declarative rendering
* Reactive state
* Component system
* Routing through Vue Router
* State management through Pinia
* Transitions/animations
* SSR/SSG capabilities
* Excellent TypeScript support

Vue can be used for:

```text
Small widget
   ↓
Single-page application
   ↓
Large frontend application
   ↓
Full-stack SSR application with Nuxt
```

### React comparison

| Vue                        | React                           |
| -------------------------- | ------------------------------- |
| Framework/library for UI   | Library for UI                  |
| `.vue` SFC                 | `.jsx/.tsx`                     |
| Template syntax            | JSX                             |
| Built-in reactivity system | State/hooks                     |
| `ref()` / `reactive()`     | `useState()`                    |
| `computed()`               | `useMemo()`                     |
| `watch()`                  | `useEffect()` (roughly)         |
| `provide/inject`           | Context                         |
| Vue Router                 | React Router / framework router |
| Pinia                      | Redux/Zustand/etc.              |
| `<Transition>`             | React Transition libraries      |
| Nuxt                       | Next.js                         |

---

# 2. Why Vue?

Vue's biggest advantage is that it gives you a lot of framework behavior while keeping the programming model relatively simple.

Important characteristics:

### Declarative

You describe **what the UI should look like**, not how to manually update the DOM.

```vue
<template>
  <h1>{{ message }}</h1>
</template>
```

Change `message`, and Vue updates the DOM automatically.

### Reactive

Vue tracks which reactive values a component uses and updates the relevant UI when they change.

### Progressive

You can use Vue:

```text
HTML enhancement
     ↓
Component library
     ↓
SPA
     ↓
SSR application
```

You don't have to adopt the entire ecosystem immediately.

---

# 3. The Progressive Framework

Vue itself focuses on the UI layer, while its official ecosystem provides:

```text
Vue
 ├── Vue Router       → routing
 ├── Pinia            → state management
 ├── Vite             → tooling/build
 └── Nuxt             → full-stack/SSR framework
```

### Next.js comparison

Think roughly:

```text
Vue + Vue Router + Pinia
        +
       Nuxt
```

is conceptually similar to the ecosystem around:

```text
React + Next.js
```

But don't assume Nuxt and Next.js have identical APIs or architecture.

---

# 4. Single-File Components (SFC)

Vue's `.vue` file can contain:

```vue
<script setup>
const name = 'John'
</script>

<template>
  <h1>Hello {{ name }}</h1>
</template>

<style scoped>
h1 {
  color: blue;
}
</style>
```

A component can therefore contain:

```text
Logic
Template
Styles
```

in one file.

### React comparison

React:

```tsx
function App() {
  const name = 'John'

  return <h1>Hello {name}</h1>
}
```

Vue:

```vue
<script setup>
const name = 'John'
</script>

<template>
  <h1>Hello {{ name }}</h1>
</template>
```

Vue separates template syntax from JavaScript instead of making HTML part of JavaScript.

---

# 5. API Styles

Vue has two API styles.

## Options API

```vue
<script>
export default {
  data() {
    return {
      count: 0
    }
  },

  computed: {
    double() {
      return this.count * 2
    }
  },

  methods: {
    increment() {
      this.count++
    }
  }
}
</script>
```

It organizes code by option:

```text
data
computed
methods
watch
mounted
props
...
```

## Composition API

```vue
<script setup>
import { ref, computed } from 'vue'

const count = ref(0)

const double = computed(() => count.value * 2)

function increment() {
  count.value++
}
</script>
```

Composition API organizes code by **feature/logic**.

### React comparison

Composition API feels much more familiar to an advanced React developer.

```text
React Hooks
      ≈
Vue Composition API
```

**Recommendation:** Learn Composition API first. Know Options API enough to maintain existing projects.

---

# 6. Creating a Vue Application

Modern Vue applications commonly start with Vite.

```bash
npm create vue@latest
cd my-app
npm install
npm run dev
```

Entry point:

```js
import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')
```

React:

```tsx
createRoot(document.getElementById('root')!).render(
  <App />
)
```

Vue's `createApp()` creates an application instance.

You can configure it:

```js
const app = createApp(App)

app.use(router)
app.use(pinia)

app.mount('#app')
```

---

# 7. Template Syntax

Vue templates look like HTML with additional directives.

```vue
<template>
  <h1>{{ title }}</h1>

  <img :src="imageUrl">

  <button @click="increment">
    Count: {{ count }}
  </button>
</template>
```

## Interpolation

```vue
{{ message }}
```

Equivalent idea:

```jsx
{message}
```

## Attribute binding

```vue
<img :src="url">
```

Long form:

```vue
<img v-bind:src="url">
```

React:

```jsx
<img src={url} />
```

## Event binding

```vue
<button @click="increment">
```

Long form:

```vue
<button v-on:click="increment">
```

React:

```jsx
<button onClick={increment}>
```

### Common Vue shorthand

```text
:foo  → v-bind:foo
@foo  → v-on:foo
#foo  → v-slot:foo
```

---

# 8. Reactivity Fundamentals

This is one of the most important Vue concepts.

## `ref()`

Use `ref()` for a reactive value.

```vue
<script setup>
import { ref } from 'vue'

const count = ref(0)

function increment() {
  count.value++
}
</script>

<template>
  <button @click="increment">
    {{ count }}
  </button>
</template>
```

Inside JavaScript:

```js
count.value
```

Inside templates:

```vue
{{ count }}
```

Vue automatically unwraps refs in templates.

### React comparison

```js
const [count, setCount] = useState(0)
```

roughly corresponds to:

```js
const count = ref(0)
```

But Vue mutates the ref:

```js
count.value++
```

instead of:

```js
setCount(count + 1)
```

---

# 9. `reactive()`

`reactive()` makes an object reactive.

```js
import { reactive } from 'vue'

const user = reactive({
  name: 'John',
  age: 25
})

user.age++
```

No `.value` is needed.

### `ref` vs `reactive`

Use:

```js
ref(0)
```

for individual values.

Use:

```js
reactive({
  name: '',
  age: 0
})
```

for object-like state.

In practice, **`ref()` is often the more flexible default**.

---

# 10. Vue Reactivity Mental Model

Vue tracks dependencies.

```js
const count = ref(0)

const double = computed(() => count.value * 2)
```

Conceptually:

```text
count
  ↓
computed
  ↓
template
  ↓
DOM
```

When `count` changes:

```text
Vue knows:
"double depends on count"
```

and updates the necessary consumers.

Vue 3's reactivity is primarily based on JavaScript Proxies for reactive objects and getter/setter tracking for refs.

---

# 11. Computed Properties

`computed()` is for **derived state**.

```js
const firstName = ref('John')
const lastName = ref('Doe')

const fullName = computed(() =>
  `${firstName.value} ${lastName.value}`
)
```

Template:

```vue
<p>{{ fullName }}</p>
```

Don't calculate expensive/derived values repeatedly in the template.

Bad:

```vue
<p>{{ users.filter(u => u.active).length }}</p>
```

Better:

```js
const activeUsers = computed(() =>
  users.value.filter(u => u.active)
)
```

### React comparison

Closest concept:

```js
useMemo(...)
```

But don't think of `computed()` as simply "Vue's useMemo".

It is primarily a **reactive derived value**.

---

# 12. Class and Style Bindings

Vue makes dynamic classes easy.

```vue
<div :class="{ active: isActive }">
```

Multiple classes:

```vue
<div :class="[baseClass, isActive ? 'active' : '']">
```

Object:

```vue
<div :class="{
  active: isActive,
  disabled: isDisabled
}">
```

Style:

```vue
<div :style="{
  color: textColor,
  fontSize: `${size}px`
}">
```

React equivalent:

```jsx
<div className={isActive ? 'active' : ''} />
```

Vue generally requires less manual class-string construction.

---

# 13. Conditional Rendering

```vue
<div v-if="isLoggedIn">
  Dashboard
</div>

<div v-else>
  Login
</div>
```

Multiple conditions:

```vue
<div v-if="status === 'loading'">Loading</div>
<div v-else-if="status === 'error'">Error</div>
<div v-else>Success</div>
```

### `v-show`

```vue
<div v-show="isVisible">
```

Difference:

```text
v-if
→ creates/removes DOM

v-show
→ keeps DOM and toggles display
```

Use `v-show` when something is toggled frequently.

React:

```jsx
{isVisible && <div />}
```

is conceptually closer to `v-if`.

---

# 14. List Rendering

```vue
<ul>
  <li v-for="user in users" :key="user.id">
    {{ user.name }}
  </li>
</ul>
```

With index:

```vue
<li v-for="(user, index) in users" :key="user.id">
```

Always provide a stable `key` when possible.

### React comparison

```jsx
users.map(user => (
  <li key={user.id}>
    {user.name}
  </li>
))
```

Vue:

```vue
<li v-for="user in users" :key="user.id">
```

The conceptual purpose is almost identical.

---

# 15. Event Handling

```vue
<button @click="increment">
  +
</button>
```

Inline:

```vue
<button @click="count++">
```

Event object:

```vue
<button @click="handleClick">
```

```js
function handleClick(event) {
  console.log(event)
}
```

## Event modifiers

Vue provides useful modifiers:

```vue
<form @submit.prevent="submit">
```

```vue
<button @click.stop="handleClick">
```

Common modifiers:

```text
.prevent
.stop
.once
.self
.capture
```

Keyboard:

```vue
<input @keyup.enter="submit">
```

This is one area where Vue templates can be considerably more concise than React.

---

# 16. Form Input Bindings

The most important form feature is `v-model`.

```vue
<input v-model="name">
```

Conceptually:

```text
value + input event
```

React usually requires:

```jsx
<input
  value={name}
  onChange={e => setName(e.target.value)}
/>
```

Vue:

```vue
<input v-model="name">
```

## Checkbox

```vue
<input type="checkbox" v-model="accepted">
```

## Select

```vue
<select v-model="country">
  <option value="de">Germany</option>
  <option value="us">USA</option>
</select>
```

Modifiers:

```vue
<input v-model.trim="name">
<input v-model.number="age">
<input v-model.lazy="value">
```

---

# 17. Watchers

`watch()` runs a side effect when reactive state changes.

```js
import { ref, watch } from 'vue'

const search = ref('')

watch(search, (newValue, oldValue) => {
  console.log(newValue)
})
```

Multiple dependencies:

```js
watch(
  [firstName, lastName],
  ([newFirst, newLast]) => {
    console.log(newFirst, newLast)
  }
)
```

Watch object:

```js
watch(
  () => user.name,
  newName => {
    console.log(newName)
  }
)
```

### When to use

Use watchers for **side effects**:

```text
API request
localStorage
analytics
imperative library
synchronizing external state
```

Don't use `watch()` just to calculate derived data.

Use:

```js
computed()
```

for derived state.

### React comparison

`watch()` is somewhat similar to:

```js
useEffect(() => {
  ...
}, [dependency])
```

But Vue's watcher is specifically designed around its reactivity system.

---

# 18. `watchEffect()`

```js
watchEffect(() => {
  console.log(user.value.name)
})
```

Vue automatically tracks the reactive values accessed inside the callback.

Conceptually:

```text
watch()
→ explicitly specify dependencies

watchEffect()
→ automatically discover dependencies
```

React's `useEffect` generally requires a dependency array.

---

# 19. Template Refs

Sometimes you need direct DOM/component access.

```vue
<script setup>
import { ref, onMounted } from 'vue'

const input = ref(null)

onMounted(() => {
  input.value.focus()
})
</script>

<template>
  <input ref="input">
</template>
```

React equivalent:

```jsx
const inputRef = useRef(null)

useEffect(() => {
  inputRef.current?.focus()
}, [])
```

Use template refs for imperative operations, not normal state management.

---

# 20. Components Basics

Parent:

```vue
<UserCard :user="user" @select="handleSelect" />
```

Child:

```vue
<script setup>
const props = defineProps({
  user: Object
})

const emit = defineEmits(['select'])
</script>

<template>
  <button @click="emit('select', user.id)">
    {{ user.name }}
  </button>
</template>
```

Component communication:

```text
Parent
  ↓ props
Child
  ↑ events
```

Exactly the important React pattern:

```text
props down
events/callbacks up
```

---

# 21. Component Registration

With `<script setup>`, imported components are automatically available.

```vue
<script setup>
import UserCard from './UserCard.vue'
</script>

<template>
  <UserCard />
</template>
```

No explicit registration required.

Traditional Options API:

```js
export default {
  components: {
    UserCard
  }
}
```

There is also global registration, but avoid overusing it because dependencies become less explicit.

---

# 22. Props

Define props:

```vue
<script setup>
const props = defineProps({
  name: String,
  age: Number
})
</script>
```

Usage:

```vue
<UserCard
  name="John"
  :age="25"
/>
```

Modern TypeScript:

```vue
<script setup lang="ts">
interface Props {
  name: string
  age: number
}

const props = defineProps<Props>()
</script>
```

Props are readonly from the child's perspective.

Don't do:

```js
props.name = 'new'
```

Instead, emit an event or maintain local derived state.

---

# 23. Events / Emits

Child:

```js
const emit = defineEmits(['save'])

emit('save', user)
```

Parent:

```vue
<UserForm @save="handleSave" />
```

```js
function handleSave(user) {
  console.log(user)
}
```

React equivalent:

```jsx
<UserForm onSave={handleSave} />
```

Vue calls these **component events**.

---

# 24. Component `v-model`

Vue supports two-way binding between components.

Parent:

```vue
<UserForm v-model="userName" />
```

Modern Vue convention means the child receives:

```text
modelValue
```

and emits:

```text
update:modelValue
```

Example:

```vue
<script setup>
defineProps({
  modelValue: String
})

const emit = defineEmits(['update:modelValue'])
</script>

<template>
  <input
    :value="modelValue"
    @input="emit('update:modelValue', $event.target.value)"
  >
</template>
```

Modern Vue also supports multiple models:

```vue
<UserForm
  v-model:first-name="firstName"
  v-model:last-name="lastName"
/>
```

### React comparison

This replaces a lot of:

```jsx
value={value}
onChange={setValue}
```

boilerplate.

---

# 25. Fallthrough Attributes

Suppose:

```vue
<MyButton class="primary" disabled />
```

and `MyButton` renders:

```vue
<button>
  Save
</button>
```

Attributes such as:

```text
class
style
id
data-*
aria-*
```

can fall through to the root element.

This is useful for reusable components.

You can control it:

```vue
<script setup>
defineOptions({
  inheritAttrs: false
})
</script>
```

Then manually bind:

```vue
<button v-bind="$attrs">
```

React developers should think of `$attrs` roughly as the concept of forwarding remaining props.

---

# 26. Slots

Slots are one of Vue's most important component composition features.

Child:

```vue
<template>
  <button class="card">
    <slot />
  </button>
</template>
```

Parent:

```vue
<Card>
  Hello
</Card>
```

Think:

```text
Vue slot ≈ React children
```

## Named slots

Child:

```vue
<header>
  <slot name="header" />
</header>

<main>
  <slot />
</main>
```

Parent:

```vue
<Card>
  <template #header>
    <h1>Hello</h1>
  </template>

  Content
</Card>
```

## Scoped slots

Child:

```vue
<slot :user="user" />
```

Parent:

```vue
<template #default="{ user }">
  {{ user.name }}
</template>
```

A useful mental model:

```text
Scoped slot ≈ render prop
```

---

# 27. Provide / Inject

Useful for avoiding prop drilling.

Provider:

```vue
<script setup>
import { provide, ref } from 'vue'

const theme = ref('dark')

provide('theme', theme)
</script>
```

Deep child:

```vue
<script setup>
import { inject } from 'vue'

const theme = inject('theme')
</script>
```

Conceptually:

```text
Vue provide/inject
        ≈
React Context
```

But it is dependency injection rather than necessarily global state.

For large applications, prefer a state manager when appropriate.

---

# 28. Async Components

Load a component lazily:

```js
import { defineAsyncComponent } from 'vue'

const UserModal = defineAsyncComponent(
  () => import('./UserModal.vue')
)
```

Useful for:

```text
Large dialogs
Rarely used pages
Heavy widgets
Code splitting
```

React equivalent:

```jsx
const UserModal = lazy(() => import('./UserModal'))
```

---

# 29. Lifecycle Hooks

Important Composition API hooks:

```js
onMounted()
onUpdated()
onUnmounted()

onBeforeMount()
onBeforeUpdate()
onBeforeUnmount()
```

Example:

```vue
<script setup>
import { onMounted, onUnmounted } from 'vue'

function handleResize() {
  console.log(window.innerWidth)
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>
```

### React comparison

```text
onMounted
    ≈ useEffect(..., [])

onUnmounted
    ≈ useEffect cleanup
```

But Vue's lifecycle is more explicitly named.

---

# 30. Composables

A **composable** is reusable reactive logic.

Example:

```js
// useCounter.js
import { ref, computed } from 'vue'

export function useCounter(initial = 0) {
  const count = ref(initial)

  const double = computed(() => count.value * 2)

  function increment() {
    count.value++
  }

  return {
    count,
    double,
    increment
  }
}
```

Use:

```vue
<script setup>
import { useCounter } from './useCounter'

const {
  count,
  double,
  increment
} = useCounter(10)
</script>
```

### React comparison

```text
Vue composable
      ≈
React custom hook
```

This is one of the easiest concepts for React developers.

---

# 31. Custom Directives

Directives manipulate DOM behavior.

Built-in:

```text
v-if
v-for
v-model
v-show
v-bind
v-on
```

Custom example:

```js
const vFocus = {
  mounted: el => el.focus()
}
```

Usage:

```vue
<input v-focus>
```

Use custom directives mainly for **low-level DOM behavior**.

Don't use them when a normal component/composable is more appropriate.

---

# 32. Plugins

A plugin extends/configures the Vue application.

```js
const myPlugin = {
  install(app) {
    app.provide('api', api)
  }
}

app.use(myPlugin)
```

Plugins are useful for:

```text
Global services
Libraries
Global configuration
Directives
Components
Dependency injection
```

Examples include router and Pinia.

---

# 33. Built-in Components

Vue includes several important built-in components.

```text
<Transition>
<TransitionGroup>
<KeepAlive>
<Teleport>
<Suspense>
```

---

# 34. Transition

For entering/leaving elements:

```vue
<Transition name="fade">
  <div v-if="visible">
    Hello
  </div>
</Transition>
```

CSS:

```css
.fade-enter-active,
.fade-leave-active {
  transition: opacity .3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
```

Vue automatically applies transition classes.

React usually needs a library or manual CSS/class management.

---

# 35. TransitionGroup

For lists:

```vue
<TransitionGroup name="list" tag="ul">
  <li
    v-for="item in items"
    :key="item.id"
  >
    {{ item.name }}
  </li>
</TransitionGroup>
```

Useful for:

```text
Add/remove
Reordering
List animations
```

---

# 36. KeepAlive

Caches component instances.

```vue
<KeepAlive>
  <component :is="currentComponent" />
</KeepAlive>
```

Without it:

```text
switch component
→ unmount
→ mount again
```

With it:

```text
switch component
→ cache
→ restore existing instance
```

Useful for tabs/forms where you want component state preserved.

React has no exact built-in equivalent.

---

# 37. Teleport

Render component content somewhere else in the DOM.

```vue
<Teleport to="body">
  <Modal />
</Teleport>
```

Very useful for:

```text
Modal
Dialog
Popover
Toast
Dropdown
```

Conceptually similar to:

```jsx
createPortal(...)
```

in React.

---

# 38. Suspense

Vue's `<Suspense>` can coordinate async dependencies.

Conceptually:

```vue
<Suspense>
  <AsyncPage />

  <template #fallback>
    Loading...
  </template>
</Suspense>
```

Think of it as related to React Suspense, but Vue's async-component handling and ecosystem patterns differ.

Use it when your application architecture actually benefits from coordinated async rendering.

---

# 39. Routing

Install Vue Router:

```bash
npm install vue-router
```

Routes:

```js
const routes = [
  {
    path: '/',
    component: Home
  },
  {
    path: '/users/:id',
    component: User
  }
]
```

Create router:

```js
const router = createRouter({
  history: createWebHistory(),
  routes
})
```

Register:

```js
app.use(router)
```

Navigation:

```vue
<RouterLink to="/users">
  Users
</RouterLink>

<RouterView />
```

Programmatic navigation:

```js
const router = useRouter()

router.push('/users')
```

Route params:

```js
const route = useRoute()

console.log(route.params.id)
```

### React comparison

```text
Vue Router
≈
React Router
```

### Next.js comparison

Next.js routing is filesystem-based.

Vue Router is explicitly configured:

```text
pages/
routes config
```

Nuxt provides file-based routing on top of Vue.

---

# 40. State Management

For small apps, local component state and composables may be enough.

For shared application state, Vue's recommended state-management solution is **Pinia**.

Example:

```js
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    user: null
  }),

  getters: {
    isLoggedIn: state => !!state.user
  },

  actions: {
    login(user) {
      this.user = user
    }
  }
})
```

Use:

```js
const userStore = useUserStore()

userStore.login(user)
```

### React comparison

```text
Pinia
≈
Redux Toolkit / Zustand
```

Pinia is generally much less boilerplate-heavy than classic Redux.

---

# 41. Single-File Components at Scale

A common project structure:

```text
src/
├── components/
├── views/
├── composables/
├── stores/
├── router/
├── services/
├── assets/
├── types/
├── App.vue
└── main.ts
```

Typical responsibility:

```text
components → reusable UI
views       → route-level UI
composables → reusable reactive logic
stores      → shared application state
services    → API/business integrations
types       → TypeScript types
```

This is conceptually similar to a well-structured React application.

---

# 42. Tooling

Modern Vue development commonly uses:

```text
Vite
Vue Devtools
TypeScript
ESLint
Vitest
Vue Test Utils
```

Vite provides:

```text
Dev server
HMR
Production build
Code splitting
Tree shaking
```

React developers will find Vite familiar.

---

# 43. Testing

Important tools:

```text
Vitest
Vue Test Utils
Playwright / Cypress
```

Test a component:

```js
import { mount } from '@vue/test-utils'

const wrapper = mount(Counter)

await wrapper.find('button').trigger('click')

expect(wrapper.text()).toContain('1')
```

Think:

```text
Vue Test Utils
≈
React Testing Library
```

Prefer testing behavior rather than implementation details.

---

# 44. SSR

Vue supports SSR, but for production applications you will often use **Nuxt** rather than building an SSR system manually.

SSR:

```text
Request
   ↓
Server renders Vue
   ↓
HTML sent to browser
   ↓
Vue hydrates
   ↓
Interactive application
```

Benefits:

```text
SEO
Initial page performance
Content delivery
```

Challenges:

```text
Hydration mismatch
Server/client differences
Data fetching
Caching
Deployment complexity
```

### Next.js comparison

```text
Nuxt
≈
Next.js
```

for the broader full-stack Vue/React framework role.

---

# 45. Production Deployment

Build:

```bash
npm run build
```

This generates a production bundle.

Typical deployment options:

```text
Static hosting
CDN
Node server
Cloud platforms
Nuxt server/runtime
```

For a pure SPA:

```text
Browser
   ↓
CDN/static files
```

For SSR:

```text
Browser
   ↓
Server/runtime
   ↓
Rendered HTML
```

---

# 46. Performance

Important Vue performance principles:

### 1. Code splitting

```js
const AdminPage = () =>
  import('./AdminPage.vue')
```

### 2. Lazy-load routes

Avoid shipping every page initially.

### 3. Stable props

Avoid unnecessarily changing props for large component trees.

### 4. Virtualize large lists

Don't render 50,000 DOM nodes if only 20 are visible.

### 5. Use computed values for derived state

```js
const filteredUsers = computed(...)
```

### 6. Avoid unnecessary deep reactivity

For large immutable structures, consider:

```js
shallowRef()
shallowReactive()
markRaw()
```

when appropriate.

### 7. Measure before optimizing

Use:

```text
Browser Performance tools
Vue Devtools
Web Vitals
```

Vue also supports optimizations such as `v-once` and `v-memo`.

---

# 47. Accessibility

Vue doesn't automatically make an application accessible.

You still need:

```text
Semantic HTML
Keyboard navigation
ARIA where necessary
Focus management
Color contrast
Labels
Accessible forms
Screen-reader support
```

Bad:

```vue
<div @click="submit">
  Submit
</div>
```

Better:

```vue
<button @click="submit">
  Submit
</button>
```

Prefer native HTML before reaching for ARIA.

---

# 48. Security

Vue automatically escapes normal template interpolation:

```vue
{{ userInput }}
```

This is safer than directly inserting HTML.

Be careful with:

```vue
<div v-html="html"></div>
```

`v-html` renders raw HTML.

Never put untrusted HTML into it without proper sanitization.

Also avoid:

```text
User-controlled Vue templates
Unsafe URLs
Untrusted JavaScript
Unsafe styles
```

Important rule:

> Vue protects you from many accidental injection problems, but it cannot make intentionally rendering untrusted HTML/JavaScript safe.

---

# 49. TypeScript with Composition API

Vue + TypeScript works particularly well with Composition API.

```vue
<script setup lang="ts">
interface User {
  id: number
  name: string
}

const user = ref<User | null>(null)
```

Props:

```ts
interface Props {
  user: User
}

const props = defineProps<Props>()
```

Events:

```ts
const emit = defineEmits<{
  save: [user: User]
  cancel: []
}>()
```

Refs:

```ts
const input = ref<HTMLInputElement | null>(null)
```

---

# 50. TypeScript with Options API

Options API also supports TypeScript:

```vue
<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  data() {
    return {
      count: 0
    }
  },

  methods: {
    increment() {
      this.count++
    }
  }
})
</script>
```

Composition API generally provides a cleaner TypeScript experience for modern Vue applications.

---

# 51. Reactivity in Depth

Understand these concepts eventually:

```text
Reactive object
Ref
Computed
Reactive effect
Dependency tracking
Effect triggering
Shallow reactivity
Readonly
Raw objects
```

Useful APIs:

```js
ref()
reactive()
computed()
readonly()
shallowRef()
shallowReactive()
toRef()
toRefs()
toValue()
unref()
markRaw()
```

Important example:

```js
const state = reactive({
  count: 0
})

const count = toRef(state, 'count')
```

`toRefs()` is useful when destructuring reactive objects without losing reactivity.

Be careful:

```js
const { count } = reactiveState
```

can lose the reactive connection.

---

# 52. Rendering Mechanism

You don't need this initially, but as an advanced developer understand the pipeline:

```text
Template
   ↓
Compiler
   ↓
Render function
   ↓
Virtual DOM
   ↓
Reactive effect
   ↓
DOM patching
```

Vue compiles templates into optimized render functions.

Vue's compiler can identify static portions and optimize updates.

### React comparison

Both use declarative rendering and a virtual DOM-based model, but Vue's compiler has more opportunities to optimize the generated rendering code because Vue controls the template compilation process.

---

# 53. Render Functions & JSX

You don't normally need render functions.

Template:

```vue
<template>
  <button>{{ label }}</button>
</template>
```

Render function:

```js
import { h } from 'vue'

export default {
  setup() {
    return () => h('button', 'Hello')
  }
}
```

JSX is also possible:

```jsx
const App = () => <button>Hello</button>
```

Use templates by default.

Use render functions/JSX when:

```text
Highly dynamic rendering
Library development
Programmatic UI
Complex rendering abstractions
```

---

# 54. Vue and Web Components

Vue can consume Web Components:

```vue
<my-button />
```

Vue can also be used to create custom elements.

Web Components are:

```text
Browser-standard custom elements
Shadow DOM
Custom events
HTML interoperability
```

Don't confuse:

```text
Vue component
```

with:

```text
Web Component
```

A Vue component is framework-level; a Web Component is a browser platform standard.

---

# 55. Animation Techniques

Vue's `<Transition>` is the normal starting point.

For more complex animations:

```text
CSS transitions
CSS animations
Transition
TransitionGroup
JavaScript animation libraries
Web Animations API
requestAnimationFrame
```

Rule:

```text
Simple UI animation
→ CSS + Transition

Complex animation
→ dedicated animation solution
```

Don't put animation logic into ordinary reactive state unless necessary.

---

# 56. Ways of Using Vue

Vue can be used in several ways.

### 1. Progressive enhancement

Add Vue to an existing page.

### 2. Standalone component

Build a reusable interactive widget.

### 3. SPA

Vue + Vue Router + Pinia.

### 4. SSR / Full-stack

Use Nuxt.

### 5. Web Components

Build custom elements.

### 6. Mobile/Desktop ecosystems

Vue can also participate in applications using frameworks such as Ionic/Capacitor, Tauri, Electron, etc.

---

# 57. Composition API FAQ — Important Points

### Why Composition API?

Because it allows related logic to stay together.

Options API:

```text
data
methods
computed
watch
```

can scatter one feature across the component.

Composition API:

```js
const users = ref([])
const loading = ref(false)

async function fetchUsers() {
  ...
}
```

keeps the feature together.

### Is Composition API required?

No.

But for modern Vue development, especially for developers coming from React, it is the API you should prioritize.

### Is Composition API just React Hooks?

No.

The programming experience is similar, but Vue's reactivity system is fundamentally different.

React:

```text
state update
→ render function runs again
```

Vue:

```text
reactive dependency changes
→ affected reactive effects/components update
```

That distinction is important.

---

# 58. The Most Important React → Vue Mapping

Memorize this table:

| React                     | Vue                         |
| ------------------------- | --------------------------- |
| JSX                       | Template                    |
| Component                 | Component                   |
| `useState()`              | `ref()` / `reactive()`      |
| `useMemo()`               | `computed()`                |
| `useEffect()`             | `watch()` / `watchEffect()` |
| `useRef()`                | Template `ref()`            |
| Custom Hook               | Composable                  |
| Context                   | Provide / Inject            |
| `children`                | Slot                        |
| Render props              | Scoped slots                |
| Callback props            | Emits                       |
| Controlled input          | `v-model`                   |
| Conditional JSX           | `v-if`                      |
| `.map()`                  | `v-for`                     |
| `className`               | `class` / `:class`          |
| `onClick`                 | `@click`                    |
| `dangerouslySetInnerHTML` | `v-html`                    |
| Portal                    | Teleport                    |
| `React.lazy`              | `defineAsyncComponent`      |
| Error/loading patterns    | Vue async/Suspense patterns |
| React Router              | Vue Router                  |
| Redux/Zustand             | Pinia                       |
| React Testing Library     | Vue Test Utils              |
| Next.js                   | Nuxt                        |
| React DevTools            | Vue Devtools                |

---

# 59. The Biggest Mental Differences

These are more important than memorizing directives.

## Difference 1 — React rerenders, Vue tracks dependencies

React:

```text
setState()
   ↓
component function runs again
```

Vue:

```text
reactive value changes
   ↓
Vue knows what depends on it
   ↓
affected UI/effects update
```

---

## Difference 2 — Vue templates are declarative HTML

React:

```jsx
return (
  <button
    className={active ? 'active' : ''}
    onClick={handleClick}
  >
    {name}
  </button>
)
```

Vue:

```vue
<button
  :class="{ active }"
  @click="handleClick"
>
  {{ name }}
</button>
```

Vue provides many template-level primitives.

---

## Difference 3 — `computed` is first-class

Instead of repeatedly calculating derived state:

```js
const filtered = computed(() =>
  users.value.filter(...)
)
```

Vue encourages derived state as a natural part of the reactive graph.

---

## Difference 4 — `v-model` is built into the framework

Instead of manually wiring:

```text
value
onChange
```

you can write:

```vue
<input v-model="name">
```

---

## Difference 5 — Components communicate through props + emits + slots

A very important Vue component model is:

```text
Props
   ↓
Child
   ↑
Emits

Slots → composition from parent
Provide/Inject → dependency passing
```

---

# 60. What You Should Actually Master

Don't spend equal time on every Vue feature.

## Tier 1 — Must Know

Master these:

```text
SFC
<script setup>
Template syntax
ref
reactive
computed
watch
watchEffect
v-if
v-for
v-model
events
props
emits
slots
component v-model
composables
lifecycle hooks
Vue Router
Pinia
TypeScript
```

These make up most day-to-day Vue development.

---

## Tier 2 — Should Know

Understand:

```text
provide/inject
async components
Teleport
KeepAlive
Transition
TransitionGroup
custom directives
plugins
template refs
fallthrough attributes
testing
performance
SSR concepts
```

---

## Tier 3 — Learn When Needed

Don't spend much time initially on:

```text
Render functions
JSX
Reactivity internals
Custom elements
Advanced compiler behavior
Suspense edge cases
Low-level rendering optimization
```

Know what they are; learn implementation details when your project requires them.

---

# 61. A Small Realistic Vue Component

This combines many important concepts:

```vue
<script setup lang="ts">
import { computed, ref } from 'vue'

interface Todo {
  id: number
  title: string
  completed: boolean
}

const todos = ref<Todo[]>([
  {
    id: 1,
    title: 'Learn Vue',
    completed: false
  }
])

const newTodo = ref('')

const remaining = computed(() =>
  todos.value.filter(todo => !todo.completed).length
)

function addTodo() {
  const title = newTodo.value.trim()

  if (!title) return

  todos.value.push({
    id: Date.now(),
    title,
    completed: false
  })

  newTodo.value = ''
}

function removeTodo(id: number) {
  todos.value = todos.value.filter(
    todo => todo.id !== id
  )
}
</script>

<template>
  <section>
    <h1>Todos</h1>

    <form @submit.prevent="addTodo">
      <input
        v-model="newTodo"
        placeholder="New todo"
      />

      <button>
        Add
      </button>
    </form>

    <p>
      Remaining: {{ remaining }}
    </p>

    <ul>
      <li
        v-for="todo in todos"
        :key="todo.id"
      >
        <label>
          <input
            v-model="todo.completed"
            type="checkbox"
          >

          <span
            :class="{ completed: todo.completed }"
          >
            {{ todo.title }}
          </span>
        </label>

        <button @click="removeTodo(todo.id)">
          Delete
        </button>
      </li>
    </ul>

    <p v-if="todos.length === 0">
      No todos.
    </p>
  </section>
</template>
```

This tiny component demonstrates:

```text
SFC
<script setup>
TypeScript
ref
computed
v-model
v-for
v-if
:key
:event handling
class binding
forms
reactivity
```

---

# 62. A Realistic Component Architecture

For a medium/large Vue application:

```text
App.vue
│
├── Router
│
├── Views
│   ├── HomeView.vue
│   ├── UsersView.vue
│   └── SettingsView.vue
│
├── Components
│   ├── UserCard.vue
│   ├── Modal.vue
│   └── DataTable.vue
│
├── Composables
│   ├── useAuth.ts
│   ├── useFetch.ts
│   └── usePagination.ts
│
├── Stores
│   ├── auth.ts
│   └── cart.ts
│
├── Services
│   ├── api.ts
│   └── userService.ts
│
└── Types
    └── user.ts
```

Think of this as roughly equivalent to a mature React application:

```text
Pages
Components
Hooks
Stores
Services
Types
```

---

# 63. Recommended Vue Learning Order for You

Since you already know advanced React:

### Phase 1 — Vue syntax

Learn:

```text
SFC
template
<script setup>
ref
reactive
computed
v-if
v-for
v-bind
v-on
v-model
```

### Phase 2 — Components

Then:

```text
props
emits
slots
component v-model
provide/inject
fallthrough attributes
```

### Phase 3 — Composition

Then:

```text
watch
watchEffect
lifecycle hooks
template refs
composables
```

### Phase 4 — Application architecture

Then:

```text
Vue Router
Pinia
API calls
TypeScript
testing
```

### Phase 5 — Advanced Vue

Finally:

```text
Teleport
KeepAlive
Transition
Suspense
async components
SSR
performance
rendering mechanism
custom directives
plugins
Web Components
```

---

# 64. Vue Rules Worth Memorizing

```text
1. Use <script setup> for modern Vue components.

2. Use ref() for reactive values.

3. Use computed() for derived state.

4. Use watch() for side effects.

5. Use composables for reusable reactive logic.

6. Use props to send data down.

7. Use emits to communicate upward.

8. Use slots for component composition.

9. Use provide/inject to avoid deep prop drilling.

10. Use Pinia for significant shared application state.

11. Use Vue Router for SPA routing.

12. Prefer native HTML for accessibility.

13. Avoid v-html with untrusted content.

14. Give v-for items stable keys.

15. Don't use watchers where computed is enough.

16. Don't manipulate the DOM directly unless necessary.

17. Don't put everything into global state.

18. Keep components focused.

19. Lazy-load large/rare features.

20. Measure performance before optimizing.
```

---

# 65. Final Mental Model

If you remember only one diagram, remember this:

```text
                 Vue Application
                       │
          ┌────────────┴────────────┐
          │                         │
      Components                Composables
          │                         │
   ┌──────┼──────┐                  │
   │      │      │                  │
 Props  Emits  Slots          Reusable Logic
   │      │      │                  │
   └──────┴──────┴──────────┬───────┘
                            │
                       Reactivity
                    ┌───────┼────────┐
                    │       │        │
                   ref   computed   watch
                    │       │        │
                    └───────┼────────┘
                            │
                         Template
                            │
                         DOM/UI
```

And at application level:

```text
Vue
│
├── Vue Router → navigation
│
├── Pinia → shared state
│
├── Composables → reusable logic
│
├── Vite → development/build
│
└── Nuxt → SSR/full-stack architecture
```

### The key transition from React

As a React expert, **do not translate every Vue feature literally into React**.

Instead, learn this Vue mindset:

```text
Reactive state
      ↓
Derived state
      ↓
Template
      ↓
Vue tracks dependencies
      ↓
Only affected UI/effects update
```

Once `ref → computed → watch → props → emits → slots → composables` becomes natural, most of Vue 3 becomes straightforward.
