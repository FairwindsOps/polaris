<template>
  <ParentLayout>
    <template #page-bottom>
      <div class="custom-footer">
        <div class="left-footer">
          <a href="https://fairwinds.com" target="_blank">Learn more about Fairwinds</a>
          <a href="https://fairwinds.com/insights" target="_blank">Try Fairwinds Insights</a>
        </div>
        <div class="right-footer">
          <a href="https://www.fairwinds.com/privacy-policy" target="_blank">Privacy Policy</a>
        </div>
      </div>
    </template>
  </ParentLayout>
</template>

<script>
import ParentLayout from '@parent-theme/layouts/Layout.vue'

export default {
  name: 'Layout',

  components: {
    ParentLayout
  }
}
</script>

<style>
  div.custom-footer {
    display: flex;
    justify-content: space-between;
    border-top: 1px solid #eaecef;
    padding: 2rem 2.5rem;
  }
  .custom-footer .left-footer {
    margin-top: 0 !important;
  }
  .custom-footer .left-footer a:first-of-type {
    margin-right: 1.5rem;
  }
  .custom-footer a {
    color: #4e6e8e;
  }
</style>
