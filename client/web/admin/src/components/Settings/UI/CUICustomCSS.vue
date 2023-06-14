<template>
  <b-card
    class="shadow-sm"
    header-bg-variant="white"
    footer-bg-variant="white"
  >
    <template #header>
      <h3 class="m-0">
        {{ $t('title') }}
      </h3>
    </template>

    <ace-editor
      data-test-id="ui-custom-css-editor"
      :font-size="14"
      :show-print-margin="true"
      :show-gutter="true"
      :highlight-active-line="true"
      width="100%"
      height="200px"
      mode="css"
      theme="chrome"
      name="editor/css"
      :on-change="v => (customCSSSettings = v)"
      :value="customCSSSettings"
      :editor-props="{
        $blockScrolling: false
      }"
    />

    <c-submit-button
      :disabled="!canManage"
      :processing="processing"
      :success="success"
      class="float-right mt-2"
      @submit="onSubmit"
    />
  </b-card>
</template>

<script>
import { Ace as AceEditor } from 'vue2-brace-editor'
import CSubmitButton from 'corteza-webapp-admin/src/components/CSubmitButton'

import 'brace/mode/css'
import 'brace/theme/chrome'

export default {
  name: 'CUIEditorCustomCSS',

  i18nOptions: {
    namespaces: 'ui.settings',
    keyPrefix: 'editor.custom-css',
  },

  components: {
    AceEditor,
    CSubmitButton,
  },

  props: {
    settings: {
      type: Object,
      required: true,
    },

    processing: {
      type: Boolean,
      value: false,
    },

    success: {
      type: Boolean,
      value: false,
    },

    canManage: {
      type: Boolean,
      required: true,
    },
  },

  data () {
    return {
      customCSSSettings: '',
    }
  },

  watch: {
    settings: {
      immediate: true,
      handler (settings) {
        this.customCSSSettings = settings['ui.custom-css'] || ''
      },
    },
  },

  methods: {
    onSubmit () {
      this.$emit('submit', { 'ui.custom-css': this.customCSSSettings })
    },
  },

}
</script>
