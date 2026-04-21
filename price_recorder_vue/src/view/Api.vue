<script setup>
import {onMounted, ref} from "vue";

const count = ref(0)
function add() {
  count.value++
}
onMounted(()=>{
  console.log("vue已经挂载完毕")
})
</script>
<template>
  <div class="max-w-[980px] mx-auto px-5 py-12">
    <h2 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 40px; font-weight: 600; line-height: 1.10; color: #1d1d1f; margin-bottom: 24px;">
      API 测试
    </h2>
    <button
      @click="add"
      class="text-white transition-colors"
      style="background-color: #0071e3; border-radius: 8px; padding: 8px 15px; font-size: 17px; font-weight: 400; line-height: 2.41;"
    >
      count: {{count}}
    </button>
    <div class="mt-6" style="font-size: 17px; letter-spacing: -0.374px; line-height: 1.47; color: rgba(0, 0, 0, 0.8);">
      测试一下
    </div>
  </div>
</template>
