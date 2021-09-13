<template>
  <q-expansion-item
    v-for="method in api"
    :key="method.path"
    group="api"
    dense
    expand-separator
    :label="method.name"
  >
    <q-card>
      <q-card-section>
        <code
          ><span class="text-light-green">{{ method.method }}</span> /api/{{
            method.path
          }}</code
        >
        <h5 class="text-caption q-mt-sm q-mb-none">Headers</h5>
        <code
          >{"X-Api-Key": "<i>{{ $store.state.wallet.adminkey }}</i
          >"}</code
        ><br />
        <code>{{ method.params }}</code>
        <h5 class="text-caption q-mt-sm q-mb-none">
          Returns 200 OK (application/json)
        </h5>
        <code>{{ method.returns }}</code>
        <h5 class="text-caption q-mt-sm q-mb-none">Curl example</h5>
        <code
          >curl {{ location.href.split('/').slice(0, 3).join('/') }}/api/{{
            method.path
          }}
          -H "X-Api-Key: <i>{{ $store.state.wallet.adminkey }}</i
          >"</code
        >
      </q-card-section>
    </q-card>
  </q-expansion-item>
</template>
