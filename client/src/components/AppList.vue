<template>
  <q-list v-if="user && apps.length" dense class="lnbits-drawer__q-list">
    <q-item-label header>Apps</q-item-label>
    <q-item
      v-for="app in $store.state.user.apps"
      :key="app.code"
      clickable
      :active="app.isActive"
      tag="a"
      :href="[app.url, '?usr=', user.id].join('')"
    >
      <q-item-section side>
        <q-avatar
          size="md"
          :color="
            app.isActive ? ($q.dark.isActive ? 'primary' : 'primary') : 'grey-5'
          "
        >
          <q-icon
            :name="app.icon"
            :size="$q.dark.isActive ? '21px' : '20px'"
            :color="$q.dark.isActive ? 'blue-grey-10' : 'grey-3'"
          ></q-icon>
        </q-avatar>
      </q-item-section>
      <q-item-section>
        <q-item-label lines="1">{{ app.name }}</q-item-label>
      </q-item-section>
      <q-item-section v-show="app.isActive" side>
        <q-icon name="chevron_right" color="grey-5" size="md"></q-icon>
      </q-item-section>
    </q-item>
    <q-item clickable tag="a" :href="['/apps?usr=', user.id].join('')">
      <q-item-section side>
        <q-icon name="clear_all" color="grey-5" size="md"></q-icon>
      </q-item-section>
      <q-item-section>
        <q-item-label lines="1" class="text-caption">Manage apps</q-item-label>
      </q-item-section>
    </q-item>
  </q-list>
</template>

<script>
export default {
  data() {
    return {}
  }
}
</script>
